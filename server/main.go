package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"sync"

	redis "gopkg.in/redis.v5"

	"github.com/bign8/cdn/health"
)

var (
	target = flag.String("target", os.Getenv("TARGET"), "target hostname")
	port   = flag.Int("port", 8081, "What port to run server on")
	cap    = flag.Int("cap", 10, "How many requests to store in cache")
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type response struct {
	code int
	head http.Header
	body []byte
}

func (r *response) Send(w http.ResponseWriter) {
	for key, value := range r.head {
		w.Header()[key] = value
	}
	w.WriteHeader(r.code)
	w.Write(r.body)
}

type cdn struct {
	rp  *httputil.ReverseProxy
	red *redis.Client
	cap int

	cache map[string]response
	mu    sync.RWMutex
}

func (c *cdn) RoundTrip(req *http.Request) (*http.Response, error) {
	log.Println("Proxying!", req.URL.String())
	res, err := http.DefaultTransport.RoundTrip(req)
	if err == nil && res.StatusCode == http.StatusOK { // TODO: trap other headers and respect cache codes
		r := response{
			code: res.StatusCode,
			head: res.Header,
		}
		r.body, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return res, err
		}
		res.Body = ioutil.NopCloser(bytes.NewReader(r.body))
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
	if ok {
		item.Send(w)
	} else {
		c.rp.ServeHTTP(w, req)
	}
}

func main() {
	health.Check()
	uri, err := url.Parse(*target)
	check(err)

	host, err := os.Hostname()
	check(err)

	red := redis.NewClient(&redis.Options{Addr: "redis:6379"}) // TODO: swap back to `redis` for compose
	check(red.Ping().Err())
	red.SAdd("cdn-servers", host)

	cdnHandler := &cdn{
		rp:    httputil.NewSingleHostReverseProxy(uri),
		cap:   *cap,
		red:   red,
		cache: make(map[string]response),
	}
	cdnHandler.rp.Transport = cdnHandler

	http.Handle("/", cdnHandler)
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("PONG")) })
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		if err := red.Incr("counter").Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		n, err := red.Get("counter").Int64()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(fmt.Sprintf("Hey! hit %d times", n)))
	})
	http.HandleFunc("/other", func(w http.ResponseWriter, r *http.Request) {
		parts, err := red.SMembers("cdn-servers").Result()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		neighbor := parts[rand.Intn(len(parts))] // TODO bign8: make sure this isn't me
		res, err := http.Get("http://" + neighbor + ":8081/ping")
		if err != nil {
			http.Error(w, err.Error(), http.StatusExpectationFailed)
			return
		}
		w.WriteHeader(res.StatusCode)
		io.Copy(w, res.Body)
		res.Body.Close()
		w.Write([]byte(fmt.Sprintf("\nMe:%s\nOther:%s", host, neighbor)))
	})
	log.Printf("ReverseProxy for %q serving on :%d\n", *target, *port)
	check(http.ListenAndServe(":"+strconv.Itoa(*port), nil))
}
