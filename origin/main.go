package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"sync/atomic"
	"text/template"
	"time"
)

const n = 1.1e4

var mask = fmt.Sprintf("%%0%dd", int(math.Log10(n)))

func pad(num int) string { return fmt.Sprintf(mask, num) }

type graph [][]int

func genGraph(numLinks int) graph {
	if numLinks > n {
		numLinks = n - 1
	}
	log.Print("Starting Graph Creation.")
	now := time.Now()
	G := make(graph, n)
	for i := range G {
		G[i] = make([]int, numLinks)
		G[i][0] = i - 1
		G[i][1] = i + 1
		for j := 2; j < numLinks; j++ {
			G[i][j] = rand.Intn(n)
		}
	}
	G[0][0] = n - 1
	G[n-1][1] = 0
	for i := range G {
		x := rand.Intn(numLinks)
		G[i][0], G[i][x] = G[i][x], G[i][0]
		x = rand.Intn(numLinks)
		G[i][1], G[i][x] = G[i][x], G[i][1]
	}
	log.Printf("Graph Completed: %s", time.Since(now))
	return G
}

const tpl = `<!DOCTYPE html>
<html>
	<head>
		<title>Page {{.Title}}</title>
	</head>
	<body>
		<h1>Page {{.Title}}</h1>
		<ul>{{range .Items}}
			<li>
				<a href="/page/{{ . }}">Link {{ . }}</a>
			</li>{{end}}
		</ul>
	</body>
</html>`

var t = template.Must(template.New("page").Parse(tpl))

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	G := genGraph(30)
	ctr := uint64(0)
	mux := http.NewServeMux()
	mux.HandleFunc("/page/", func(w http.ResponseWriter, r *http.Request) {
		// now := time.Now()
		u, err := strconv.Atoi(r.URL.Path[6:])
		// u, err := strconv.ParseUint(r.URL.Path[6:], 10, 64) // 6 = len("/page/")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		if u > n || u < 0 {
			http.Redirect(w, r, "/page/"+pad(0), http.StatusTemporaryRedirect)
			return
		}

		links := make([]string, len(G[u]))
		for i, n := range G[u] {
			links[i] = pad(n)
		}

		data := struct {
			Title string
			Items []string
		}{pad(u), links}
		if err := t.Execute(w, data); err != nil {
			log.Fatal(err)
		}
		// log.Printf("%s %s", r.URL.Path, time.Since(now))
		atomic.AddUint64(&ctr, 1)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/page/"+pad(0), http.StatusTemporaryRedirect)
	})
	go func() {
		total := uint64(0)
		for {
			val := atomic.LoadUint64(&ctr)
			if val != 0 {
				total += val
				fmt.Printf("QPS: %d; Total: %d\n", val, total)
				atomic.StoreUint64(&ctr, 0)
			}
			time.Sleep(time.Second)
		}
	}()
	http.ListenAndServe(":8080", mux)
}
