package health

import (
	"flag"
	"net/http"
	"os"
)

var (
	// Version is bound by the Makefile to be the version of our build
	Version = "Unknown"

	hc    = flag.String("hc", "", "Should we run a health check?")
	get   = http.DefaultClient.Get
	exit  = os.Exit
	write = os.Stdout.WriteString
)

// Static creates a static http response handler
func Static(msg string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(msg))
	}
}

// Check verifies a particular services give a 200 hc response
func Check() {
	flag.Parse()
	http.DefaultServeMux = &http.ServeMux{} // Resets previous bindings
	http.HandleFunc("/ping", Static("PONG"))
	http.HandleFunc("/version", Static(Version))
	if *hc == "" {
		return
	} else if res, err := get(*hc); err != nil {
		write("err:" + err.Error())
		exit(1)
	} else if res.StatusCode != http.StatusOK {
		write("status:" + res.Status)
		exit(1)
	} else {
		res.Body.Close()
		write("OK")
		exit(0)
	}
}
