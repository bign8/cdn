package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"

	metrics "github.com/rcrowley/go-metrics"
	boom "github.com/tylertreat/BoomFilters"
	redis "gopkg.in/redis.v5"

	"github.com/bign8/cdn/server/DHT"
)

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

func (c *cdn) RoundTrip(req *http.Request) (*http.Response, error) {
	res, err := http.DefaultTransport.RoundTrip(req)
	if err == nil && res.StatusCode == http.StatusOK { // TODO: trap other headers and respect cache codes
		var r response
		r, err = newResponse(res)
		if err != nil {
			log.Print(c.me + " problem creating newResponse " + err.Error())
			return res, err
		}
		c.mu.Lock()
		c.cache[req.URL.Path] = r
		c.cacheSize.Update(int64(len(c.cache)))
		c.bloom.Add([]byte(req.URL.Path))
		c.mu.Unlock()
	}
	return res, err
}

func (c *cdn) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	now := time.Now()
	c.mu.RLock()
	item, ok := c.cache[req.URL.Path] // TODO: respect cache timeouts
	c.mu.RUnlock()

	if ok {
		// Owns data, has data, and sending it back.
		item.Send(w)
		log.Print(c.me + " owns data and sending back")
	} else {
		serverName := c.dht.Who(req.URL.Path)
		// I own the data and don't have it, getting from origin
		if serverName == c.me {
			log.Print(c.me + " owns data and getting it from origin")
			c.rp.ServeHTTP(w, req) // Couldn't find it anywhere, sending to origin
		} else { // Send it to the true owner
			log.Print(c.me+" forwarding req onto owner: ", serverName)
			result, _ := c.DHTFetch(req.URL.Path, serverName)
			result.Send(w)
		}
	}

	c.requests.UpdateSince(now) // TODO: toggle which timer based on the branch of the redirect tree above
}
