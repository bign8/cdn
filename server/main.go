package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
)

var (
	target = flag.String("target", os.Getenv("TARGET"), "target hostname")
	port   = flag.Int("port", 8081, "What port to run server on")
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type cdn struct{}

func (c *cdn) RoundTrip(req *http.Request) (*http.Response, error) {
	log.Println("Proxying!", req.URL.String())
	return nil, http.ErrSkipAltProtocol
}

func main() {
	flag.Parse()
	uri, err := url.Parse(*target)
	check(err)
	rp := httputil.NewSingleHostReverseProxy(uri)
	CDN := &cdn{}

	// Some hackery to create a man-in-the middle proxy
	tp := http.DefaultTransport.(*http.Transport)
	tp.RegisterProtocol("http", CDN)
	rp.Transport = tp

	http.Handle("/", rp)
	log.Printf("ReverseProxy for %q serving on :%d\n", *target, *port)
	check(http.ListenAndServe(":"+strconv.Itoa(*port), nil))
}
