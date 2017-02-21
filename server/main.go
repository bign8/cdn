package main

import (
	"flag"
	"fmt"
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

func main() {
	uri, err := url.Parse(*target)
	check(err)
	rp := httputil.NewSingleHostReverseProxy(uri)
	http.Handle("/", rp)
	fmt.Printf("ReverseProxy for %q serving on :%d\n", *target, *port)
	check(http.ListenAndServe(":"+strconv.Itoa(*port), nil))
}
