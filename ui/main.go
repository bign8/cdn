package main

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"

	"golang.org/x/net/websocket"

	"github.com/bign8/cdn/util/health"
)

var port = flag.Int("port", 8083, "what port to run server on")

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
	http.HandleFunc("/hello", man.Hello)
	http.Handle("/ws/", websocket.Handler(man.Handle))
	http.Handle("/", http.FileServer(http.Dir("./src")))

	// Should listen to docker port mappings to ping each container directly
	// Also should have some sweet gopherjs transpiled stockets stuff for status updates
	fmt.Println("TODO: build a sweet UI that pulls vars and container healthchecks")
	http.ListenAndServe(":"+strconv.Itoa(*port), nil)
}
