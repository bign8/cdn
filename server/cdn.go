package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"sync"

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

	cache map[string]response
	mu    sync.RWMutex

	ring   []string
	ringMu sync.RWMutex
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
		c.mu.Unlock()
	}
	return res, err
}

func (c *cdn) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c.mu.RLock()
	item, ok := c.cache[req.URL.Path] // TODO: respect cache timeouts
	c.mu.RUnlock()

	if ok { // We have the data!!!
		item.Send(w)
	} else if req.Header.Get(cdnHeader) != "" {
		log.Print(c.me + " couldn't find response for neighbor")
		http.NotFound(w, req) // Request was from other CDN server, don't ask others or origin
	} else if item, ok = c.checkNeighbors(req.URL.Path); ok {
		item.Send(w) // Found request on neighbor, sending response
	} else {
		c.rp.ServeHTTP(w, req) // Couldn't find it anywhere, sending to origin
	}
}
