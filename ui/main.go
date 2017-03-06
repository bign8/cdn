package main

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"

	"github.com/bign8/cdn/util/health"
)

var port = flag.Int("port", 8083, "what port to run server on")

func main() {
	health.Check()
	// Should listen to docker port mappings to ping each container directly
	// Also should have some sweet gopherjs transpiled stockets stuff for status updates
	fmt.Println("TODO: build a sweet UI that pulls vars and container healthchecks")
	http.ListenAndServe(":"+strconv.Itoa(*port), nil)
}
