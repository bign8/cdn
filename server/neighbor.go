package main

import (
	"errors"
	"log"
	"net/http"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"
)

type neighborResult struct {
	res response
	err error
}

func (c *cdn) DHTFetch(path string, owner string) (result response, found bool) {
	log.Print("Making a DHT fetch")
	target := "http://" + owner + ":" + strconv.Itoa(*port) + path
	log.Print("DHT target: ", target)
	var r neighborResult
	if req, err := http.NewRequest(http.MethodGet, target, nil); err != nil {
		r.err = err
	} else {
		req.Header.Set(cdnHeader, c.me)
		if res, err := http.DefaultClient.Do(req); err != nil {
			r.err = err
		} else if res.StatusCode == http.StatusOK {
			r.res, r.err = newResponse(res)
		} else {
			r.err = errors.New("fetch: bad response: " + res.Status)
		}
	}
	return r.res, r.err == nil
}

func (c *cdn) monitorNeighbors() {
	var last string
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovering from catastophic ERROR in server.", r)
			debug.PrintStack()
		}
	}()
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
		c.dht.Update(result[:])

		// Wait for another cycle // TODO: listen to pub-sub for updates or something
		time.Sleep(time.Second * 5)
	}
}
