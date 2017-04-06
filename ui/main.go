package main

import (
	"flag"
	"net/http"
	"os"
	"strconv"
	"time"

	"golang.org/x/net/websocket"

	"github.com/bign8/cdn/util/health"
)

var (
	port  = flag.Int("port", 8083, "what port to run server on")
	start = time.Now()
	last  = time.Now()
	host  = "unknown"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	health.Check()

	// create manager
	man, err := newManager()
	check(err)
	host, err = os.Hostname()
	check(err)
	http.HandleFunc("/reset", man.Reset)
	http.HandleFunc("/hello", man.Hello)
	http.HandleFunc("/data", man.Data)
	http.Handle("/ws/", websocket.Handler(man.Handle))
	http.Handle("/", http.FileServer(http.Dir("./src")))

	go man.poll()

	// Should listen to docker port mappings to ping each container directly
	// Also should have some sweet gopherjs transpiled stockets stuff for status updates
	http.ListenAndServe(":"+strconv.Itoa(*port), nil)
}
