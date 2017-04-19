package main

import (
	"context"
	"encoding/base64"
	"errors"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	boom "github.com/tylertreat/BoomFilters"
)

type neighborResult struct {
	res response
	err error
}

func (c *cdn) checkNeighbors(path string) (result response, found bool) {
	c.ringMu.RLock()
	neighbors := c.ring[:]
	c.ringMu.RUnlock()
	ctx, done := context.WithTimeout(context.Background(), time.Second*5)

	// Parallel fetching function
	fetch := func(n string, fin chan<- neighborResult) {
		c.s2scalls.Inc(1)
		target := "http://" + n + ":" + strconv.Itoa(*port) + path
		var r neighborResult
		if req, err := http.NewRequest(http.MethodGet, target, nil); err != nil {
			r.err = err
		} else {
			req = req.WithContext(ctx)
			req.Header.Set(cdnHeader, c.me)
			if res, err := http.DefaultClient.Do(req); err != nil {
				r.err = err
			} else if res.StatusCode == http.StatusOK {
				r.res, r.err = newResponse(res)
			} else {
				r.err = errors.New("fetch: bad response: " + res.Status)
			}
		}
		fin <- r
	}

	// Fetch requests in paralell
	results := make(chan neighborResult, len(neighbors))
	for _, neighbor := range neighbors {
		go fetch(neighbor, results)
	}

	// fetch all results until found
	for i := 0; i < len(neighbors); i++ {
		back := <-results
		if !found && back.err == nil {
			c.nHit.Inc(1)
			log.Printf("%s: Found response on neighbor for %q", c.me, path)
			done()
			found = true
			result = back.res
		} else if !found && back.err != nil {
			c.nMiss.Inc(1)
			log.Printf("%s: Problem fetching %q from neighbor %s", c.me, path, back.err)
		}
	}
	done()
	return result, found
}

func (c *cdn) monitorNeighbors() {
	var last string
	for {
		// Get set from redis
		servers, err := c.red.SMembers("cdn-servers").Result()
		if err != nil {
			log.Print(c.me + " Cannot fetch neighbor set: " + err.Error())
			continue
		}

		// Generate usable list for consumers
		result := make([]string, 0, len(servers)-1)
		for _, server := range servers {
			if server != c.me {
				result = append(result, server)
			}
		}
		sort.Strings(result)

		// Use string representation of neighbors to determine if update is necessary
		if next := strings.Join(result, ", "); next != last {
			log.Print(c.me + ": Updating server list: [" + next + "]")
			last = next
			c.ringMu.Lock()
			c.ring = result
			c.ringMu.Unlock()
		}

		// Wait for another cycle // TODO: listen to pub-sub for updates or something
		time.Sleep(time.Second * 5)
	}
}

const plen = len("cdn.server.bloom.")

func (c *cdn) recvUpdates() {
	for {
		msg, err := c.ps.ReceiveMessage()
		if err != nil {
			log.Println("monitorNeighborsFilters: error 1", err)
			continue
		}
		neighbor := msg.Channel[plen:]
		if neighbor == c.me {
			continue // I'm talking to myself!
		}
		log.Println("monitorNeighborsFilters message:", neighbor, len(msg.Payload))
		bits, err := base64.StdEncoding.DecodeString(msg.Payload)
		if err != nil {
			log.Println("monitorNeighborsFilters: error 2", err)
			continue
		}
		c.ringMu.RLock()
		obj, ok := c.state[neighbor]
		c.ringMu.RUnlock()
		if !ok {
			obj = new(boom.BloomFilter)
			c.ringMu.Lock()
			c.state[neighbor] = obj
			c.ringMu.Unlock()
		}
		if err = obj.GobDecode(bits); err != nil {
			log.Println("monitorNeighborsFilters: error 3", err)
		}
	}
}

func (c *cdn) sendUpdates() {
	for range time.Tick(15 * time.Second) {
		c.mu.RLock()
		bits, err := c.bloom.GobEncode()
		c.mu.RUnlock()
		if err == nil {
			c.red.Publish("cdn.server.bloom."+c.me, base64.StdEncoding.EncodeToString(bits))
		} else {
			log.Println("Problem serializing BOOM!", err)
		}
	}
}
