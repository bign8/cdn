package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"

	"golang.org/x/net/html"
)

var target = flag.String("target", os.Getenv("TARGET"), "target hostname")

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()
	if *target == "" {
		fmt.Println("target is required")
		os.Exit(1)
	}

	// TODO: use an optimal queue here instead of an array
	queue := []string{*target}
	for len(queue) > 0 {
		fmt.Println("Browsing to", queue[0])
		res, err := http.Get(queue[0])
		check(err)
		queue = queue[1:]

		links := getLinks(res)

		// Browse to random link on page
		queue = append(queue, links[rand.Intn(len(links))])
		time.Sleep(time.Duration(rand.Int63n(int64(time.Second))))
	}
}

func getLinks(res *http.Response) (links []string) {
	z := html.NewTokenizer(res.Body)
	defer res.Body.Close()

	for {
		switch tt := z.Next(); tt {
		case html.ErrorToken:
			return // End of the document, we're done
		case html.StartTagToken:
			t := z.Token()
			if t.Data == "a" {
				for _, a := range t.Attr {
					if a.Key == "href" {
						link, err := url.Parse(a.Val)
						check(err)
						links = append(links, res.Request.URL.ResolveReference(link).String())
						break
					}
				}
			}
		}
	}
}
