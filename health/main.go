package main

import (
	"flag"
	"net/http"
	"os"
)

var target = flag.String("target", os.Getenv("TARGET"), "target hostname")

func main() {
	flag.Parse()
	res, err := http.Get(*target)
	if err != nil {
		os.Stdout.WriteString("err:" + err.Error() + "\n")
		os.Exit(1)
	} else if res.StatusCode != http.StatusOK {
		os.Stdout.WriteString("status:" + res.Status)
		os.Exit(1)
	}
	res.Body.Close()
	os.Stdout.WriteString("OK")
	os.Exit(0)
}
