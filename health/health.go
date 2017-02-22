package health

import (
	"flag"
	"net/http"
	"os"
)

var hc = flag.String("hc", "", "Should we run a health check?")

// Check verifies a particular services give a 200 hc response
func Check() {
	flag.Parse()
	if *hc == "" {
		return
	}
	res, err := http.Get(*hc)
	if err != nil {
		os.Stdout.WriteString("err:" + err.Error())
		os.Exit(1)
	} else if res.StatusCode != http.StatusOK {
		os.Stdout.WriteString("status:" + res.Status)
		os.Exit(1)
	}
	res.Body.Close()
	os.Stdout.WriteString("OK")
	os.Exit(0)
}
