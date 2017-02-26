package health

import (
	"flag"
	"net/http"
	"os"
)

var (
	hc    = flag.String("hc", "", "Should we run a health check?")
	get   = http.DefaultClient.Get
	exit  = os.Exit
	write = os.Stdout.WriteString
)

// Check verifies a particular services give a 200 hc response
func Check() {
	flag.Parse()
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
