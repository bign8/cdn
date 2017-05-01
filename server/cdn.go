package main

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"
	"sync"
	"time"

	metrics "github.com/rcrowley/go-metrics"
	boom "github.com/tylertreat/BoomFilters"
	redis "gopkg.in/redis.v5"

	"github.com/bign8/cdn/server/DHT"
)

type contextKey string

const contextKeyOwner contextKey = "owner"

var (
	_ http.Handler      = (*cdn)(nil)
	_ http.RoundTripper = (*cdn)(nil)
)

type cdn struct {
	me  string
	rp  *httputil.ReverseProxy
	red *redis.Client
	cap int
	ps  *redis.PubSub

	// TODO(bign8): abstract these vars to be part of a better cache object
	// Based off of the second talk in https://www.bigmarker.com/remote-meetup-go/GoSF-Go-Project-Structure-Concurrent-Data-Structures-and-Libraries
	bloom *boom.BloomFilter
	cache map[string]response
	mu    sync.RWMutex

	// TODO(bign8): abstract this to a more logical server ring object
	state  map[string]*boom.BloomFilter
	ring   []string
	ringMu sync.RWMutex
	dht    DHT.DHT

	// stats
	cacheSize metrics.Gauge
	requests  metrics.Timer
	s2scalls  metrics.Counter
	nHit      metrics.Counter
	nMiss     metrics.Counter
}

func (c *cdn) fromNeighbor(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		c.ServeHTTP(w, req)
		return
	}
	what := req.Header.Get(cdnHeader)
	req.Header.Del(cdnHeader)
	defer req.Body.Close()
	bits, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.mu.Lock()
	c.cache[what] = response{
		code: http.StatusOK, // HACK: because we only cache 200s
		head: req.Header,
		body: bits,
	}
	c.cacheSize.Update(int64(len(c.cache)))
	c.bloom.Add([]byte(what))
	c.mu.Unlock()

	log.Print(c.me, " accepting cache content for ", what, " from neighbor")
	http.Error(w, "THANKS!", http.StatusAccepted)
}

func (c *cdn) RoundTrip(req *http.Request) (*http.Response, error) {
	res, err := http.DefaultTransport.RoundTrip(req)
	if err == nil && res.StatusCode == http.StatusOK { // TODO: trap other headers and respect cache codes
		var r response
		r, err = newResponse(res)
		if err != nil {
			log.Print(c.me + " problem creating newResponse " + err.Error())
			return res, err
		}

		if addr := req.Context().Value(contextKeyOwner).(string); addr == c.me {
			// I am the rightful owner of this content, add to my cache
			c.mu.Lock()
			c.cache[req.URL.Path] = r
			c.cacheSize.Update(int64(len(c.cache)))
			c.bloom.Add([]byte(req.URL.Path))
			c.mu.Unlock()
		} else {
			// I fetched the context for a peer, send to them accordingly
			c.postToNeighbors(r, req.URL.Path, addr)
		}
	}
	return res, err
}

func (c *cdn) postToNeighbors(r response, path, who string) {
	log.Print(c.me, " is pushing ", path, " to neighbor ", who, "TODO")
	req, err := http.NewRequest(http.MethodPost, "http://"+who+":"+strconv.Itoa(*port)+"/2neighbor", bytes.NewReader(r.body))
	if err != nil {
		log.Print(c.me, " problem creating request ", err.Error())
		return
	}
	for key := range r.head {
		req.Header.Set(key, r.head.Get(key))
	}
	req.Header.Add(cdnHeader, req.URL.Path)
	res, err := http.DefaultClient.Do(req)
	if err == nil && res.StatusCode != http.StatusAccepted {
		err = errors.New("no-good response code: " + res.Status)
	}
	if err != nil {
		log.Print(c.me, " problem sending neighbor data (swallowed): ", err.Error())
		return
	}
	log.Print(c.me, " successfully sent ", path, " to neighbor ", who)
}

func (c *cdn) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer c.requests.UpdateSince(time.Now()) // TODO: toggle which timer based on the branch of the redirect tree above

	c.ringMu.RLock()
	who := c.dht.Who(req.URL.Path)
	c.ringMu.RUnlock()

	// Inject the rightful owner onto the request context for use if rp.ServeHTTP gets request
	req = req.WithContext(context.WithValue(req.Context(), contextKeyOwner, who))

	if who == c.me { // I own that data!
		c.mu.RLock()
		item, ok := c.cache[req.URL.Path] // TODO: respect cache timeouts
		c.mu.RUnlock()
		if !ok {
			log.Print(c.me + " owns data and getting it from origin")
			c.rp.ServeHTTP(w, req)
		} else {
			log.Print(c.me + " owns data and sending back")
			item.Send(w)
		}
		return
	}

	// I don't own the data, figure out if other has the data
	c.ringMu.RLock()
	bloom, ok := c.state[who]
	c.ringMu.RUnlock()

	// We think the neighbor has it, as them for it!
	if ok && bloom.Test([]byte(req.URL.Path)) {
		log.Print(c.me, " forwarding req onto owner: ", who)
		res, err := c.DHTFetch(req.URL.Path, who)
		if err == nil {
			c.nHit.Inc(1)
			res.Send(w)
			return // Successfull neighbor check!
		}
		c.nMiss.Inc(1)
		log.Print(c.me, " problem fetching ", req.URL.Path, " from ", who)
	}

	// Neighbor didn't have it or we ran into another issue, requesting from origin
	c.rp.ServeHTTP(w, req)
}
