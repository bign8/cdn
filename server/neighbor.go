package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
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

	//check specific neighbor according to DHT
	//if me, pass back, else forward

	// Parallel fetching function
	fetch := func(n string, fin chan<- neighborResult) {
		target := "http://" + n + ":" + strconv.Itoa(*port) + path
		c.dht.Who(target)
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
			log.Print(c.me + " Found response on neighbor")
			done()
			found = true
			result = back.res
		} else if !found && back.err != nil {
			log.Print(c.me + " Problem fetching from neighbor " + back.err.Error())
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
			log.Print(c.me + " is updating server list: [" + next + "]")
			last = next
			c.ringMu.Lock()
			c.ring = result
			c.ringMu.Unlock()
		}
		c.dht.Update(result[:])

		// Wait for another cycle // TODO: listen to pub-sub for updates or something
		time.Sleep(time.Second * 5)
	}
}
