package main

import (
	_ "expvar"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"golang.org/x/net/html"

	"github.com/bign8/cdn/util/health"
	"github.com/bign8/cdn/util/stats"
	metrics "github.com/rcrowley/go-metrics"
)

var (
	target = flag.String("target", os.Getenv("TARGET"), "target hostname")
	delay  = flag.Duration("delay", time.Second, "delay between page views")

	timer  metrics.Timer
	render metrics.Timer
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	health.Check()
	if *target == "" {
		fmt.Println("target is required")
		os.Exit(1)
	}
	go http.ListenAndServe(":8082", nil)

	host, err := os.Hostname()
	check(err)
	registry := stats.New("client", host)
	timer = registry.Timer("request")
	render = registry.Timer("render")

	next := *target
	for {
		fmt.Println("Browsing to", next) // TODO: wrap this into nice stats wrapper
		links := load(next)
		next = links[rand.Intn(len(links))]
		time.Sleep(time.Duration(rand.Int63n(int64(*delay))))
	}
}

func load(loc string) []string {
	start := time.Now()
	res := timeGet(loc)
	parts, links := parse(res)
	var wg sync.WaitGroup
	wg.Add(len(parts))
	for _, part := range parts {
		go func(p string) {
			timeGet(p).Body.Close()
			wg.Done()
		}(part)
	}
	wg.Wait()
	render.UpdateSince(start)
	log.Printf("Rendering %q took %s", loc, time.Since(start))
	return links
}

func timeGet(loc string) *http.Response {
	start := time.Now()
	res, err := http.Get(loc)
	timer.UpdateSince(start)
	log.Printf("Loading %q took %s", loc, time.Since(start))
	check(err)
	return res
}

func parse(res *http.Response) (resources, links []string) {
	z := html.NewTokenizer(res.Body)
	defer res.Body.Close()

	for {
		switch tt := z.Next(); tt {
		case html.ErrorToken:
			return // End of the document, we're done
		case html.StartTagToken:
			t := z.Token()
			switch t.Data {
			case "a": // link
				mine(res, t, &links, "href")
			case "img": // img
				fallthrough
			case "script": // js
				mine(res, t, &resources, "src")
			case "link": // css
				mine(res, t, &resources, "href")
			}
		}
	}
}

func mine(res *http.Response, t html.Token, list *[]string, attr string) {
	for _, a := range t.Attr {
		if a.Key == attr {
			link, err := url.Parse(a.Val)
			check(err)
			*list = append(*list, res.Request.URL.ResolveReference(link).String())
			break
		}
	}
}
