package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"

	redis "gopkg.in/redis.v5"

	"github.com/bign8/cdn/server/DHT"
	"github.com/bign8/cdn/util/health"
)

const cdnHeader = "x-bign8-cdn"

var (
	target = flag.String("target", os.Getenv("TARGET"), "target hostname")
	port   = flag.Int("port", 8081, "What port to run server on")
	cap    = flag.Int("cap", 20, "How many requests to store in cache")
)

//TODO: better fun error handlings
func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	health.Check()
	uri, err := url.Parse(*target)
	check(err)

	host, err := os.Hostname()
	check(err)
	if os.Getenv("HOST") != "" {
		host = os.Getenv("HOST")
	}

	// Localhost for local redis server (screenshot), redis for docker compose (./run.sh server in /cdn)
	// red := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	red := redis.NewClient(&redis.Options{Addr: "redis:6379"})
	check(red.Ping().Err())
	red.SAdd("cdn-servers", host)

	cdnHandler := &cdn{
		me:    host,
		rp:    httputil.NewSingleHostReverseProxy(uri),
		cap:   *cap,
		red:   red,
		cache: make(map[string]response),
		dht: &DHT.SimplisticDHT{
			DataMap: make(map[int]string),
			MyName:  host,
		},
	}
	cdnHandler.rp.Transport = cdnHandler
	http.Handle("/", cdnHandler)

	// Actually start the server
	log.Printf("ReverseProxy for %q serving on :%d\n", *target, *port)
	go cdnHandler.monitorNeighbors()

	check(http.ListenAndServe(":"+strconv.Itoa(*port), nil))
}
