package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"

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

type cdn struct {
	cap int
}

func (c *cdn) RoundTrip(req *http.Request) (*http.Response, error) {
	log.Println("Proxying!", req.URL.String())
	return http.DefaultTransport.RoundTrip(req)
}

func main() {
	health.Check()
	uri, err := url.Parse(*target)
	check(err)
	rp := httputil.NewSingleHostReverseProxy(uri)
	rp.Transport = &cdn{
		cap: *cap,
	}
	http.Handle("/", rp)
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("OK")) })
	log.Printf("ReverseProxy for %q serving on :%d\n", *target, *port)
	check(http.ListenAndServe(":"+strconv.Itoa(*port), nil))
}
