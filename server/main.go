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
	"github.com/bign8/cdn/util/stats"
	boom "github.com/tylertreat/BoomFilters"
)

const cdnHeader = "x-bign8-cdn"

var (
	target = flag.String("target", os.Getenv("TARGET"), "target hostname")
	port   = flag.Int("port", 8081, "What port to run server on")
	cap    = flag.Int("cap", 20, "How many requests to store in cache")
)

//TODO (bign8): better fun error handlings
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

	// Localhost for local redis server, redis for docker compose
	// red := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	registry := stats.New("server", host, *port)

	opts, err := redis.ParseURL("redis://" + os.Getenv("REDIS"))
	check(err)
	red := redis.NewClient(opts) //&redis.Options{Addr: "localhost:6379"})
	check(red.Ping().Err())
	red.SAdd("cdn-servers", host)

	pubsub, err := red.PSubscribe("cdn.server.bloom.*")
	check(err)
	defer pubsub.Close()

	cdnHandler := &cdn{
		ps:    pubsub,
		me:    host,
		rp:    httputil.NewSingleHostReverseProxy(uri),
		cap:   *cap,
		red:   red,
		cache: make(map[string]response),
		bloom: boom.NewBloomFilter(1000, 0.01),
		state: make(map[string]*boom.BloomFilter, 3), // MAGIC-NUMBER(3): close to the number of servers in cluster
		dht:   DHT.NewDHT(host),

		// stats objects
		cacheSize: registry.Gauge("cacheSize"),
		requests:  registry.Timer("requests"),
		s2scalls:  registry.Counter("s2s_calls"),
		nHit:      registry.Counter("neighbor_hit"),
		nMiss:     registry.Counter("neighbor_miss"),
		fPush:     registry.Counter("force_push"),
	}

	cdnHandler.rp.Transport = cdnHandler
	http.HandleFunc("/2neighbor", cdnHandler.fromNeighbor) // TODO (bign8): make this more RESTful
	http.Handle("/", cdnHandler)

	// Actually start the server
	log.Printf(host+": ReverseProxy for %q serving on :%d\n", *target, *port)
	go cdnHandler.monitorNeighbors()
	go cdnHandler.recvUpdates()
	go cdnHandler.sendUpdates()
	check(http.ListenAndServe(":"+strconv.Itoa(*port), nil))
}
