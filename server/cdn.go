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
			return res, err
		}
		c.mu.Lock()
		c.cache[req.URL.Path] = r
		c.cacheSize.Update(int64(len(c.cache)))
		c.mu.Unlock()
	}
	return res, err
}

func (c *cdn) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	now := time.Now()
	c.mu.RLock()
	item, ok := c.cache[req.URL.Path] // TODO: respect cache timeouts
	c.mu.RUnlock()

	if ok { // We have the data!!!
		log.Printf("%s: We have the data for %q", c.me, req.URL.Path)
		item.Send(w)
	} else if req.Header.Get(cdnHeader) != "" {
		log.Print(c.me + ": Couldn't find response for neighbor: " + req.URL.Path)
		http.NotFound(w, req) // Request was from other CDN server, don't ask others or origin
	} else if item, ok = c.checkNeighbors(req.URL.Path); ok {
		item.Send(w) // Found request on neighbor, sending response
	} else {
		c.rp.ServeHTTP(w, req) // Couldn't find it anywhere, sending to origin
	}
	c.requests.UpdateSince(now) // TODO: toggle which timer based on the branch of the redirect tree above
}
