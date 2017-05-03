package main

import (
	"encoding/base64"
	"errors"
	"hash/fnv"
	"log"
	"net/http"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"

	boom "github.com/tylertreat/BoomFilters"
)

func (c *cdn) DHTFetch(path string, owner string) (result response, err error) {
	c.s2scalls.Inc(1)
	target := "http://" + owner + ":" + strconv.Itoa(*port) + path
	log.Print("DHT target: ", target)
	var (
		req *http.Request
		res *http.Response
	)
	if req, err = http.NewRequest(http.MethodGet, target, nil); err == nil {
		// req.Header.Set(cdnHeader, c.me)
		if res, err = http.DefaultClient.Do(req); err == nil && res.StatusCode == http.StatusOK {
			result, err = newResponse(res)
		} else if err == nil {
			err = errors.New("fetch: bad response: " + res.Status)
		}
	}
	return result, err
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

		// Wait for another cycle // TODO (bign8): listen to pub-sub for updates or something
		time.Sleep(time.Second * 5)
	}
}

const (
	prefix = "cdn.server.bloom."
	prelen = len(prefix)
)

func (c *cdn) recvUpdates() {
	for {
		msg, err := c.ps.ReceiveMessage()
		if err != nil {
			log.Println("monitorNeighborsFilters: error 1", err)
			continue
		}
		neighbor := msg.Channel[prelen:]
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
			obj.SetHash(fnv.New64()) // so we don't get Null pointers
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
	var last string
	for range time.Tick(15 * time.Second) {
		c.mu.RLock()
		bits, err := c.bloom.GobEncode()
		c.mu.RUnlock()
		if err == nil {
			next := base64.StdEncoding.EncodeToString(bits)
			if next != last {
				_, err = c.red.Publish(prefix+c.me, next).Result()
				if err == nil {
					last = next
				} else {
					log.Println("Problem sending", err)
				}
			}
		} else {
			log.Println("Problem serializing BOOM!", err)
		}
	}
}
