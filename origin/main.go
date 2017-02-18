package main

import (
	"flag"
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

var (
	n    = flag.Int("size", 1100, "How many nodes to build website")
	l    = flag.Int("link", 30, "How many links on each page")
	seed = flag.Int64("seed", 0, "Seed of the random package")
	port = flag.Int("port", 8080, "What port to run server on")

	// TODO: fix bug with padding (try n=1100)
	mask = fmt.Sprintf("%%0%dd", int(math.Log10(float64(*n))))
)

func pad(num int) string { return fmt.Sprintf(mask, num) }

type graph [][]int

func genGraph(numLinks int) graph {
	rander := rand.New(rand.NewSource(*seed))
	if numLinks > *n {
		numLinks = *n - 1
	}
	log.Print("Starting Graph Creation.")
	now := time.Now()
	G := make(graph, *n)
	for i := range G {
		G[i] = make([]int, numLinks)
		G[i][0] = i - 1
		G[i][1] = i + 1
		for j := 2; j < numLinks; j++ {
			G[i][j] = rander.Intn(*n)
		}
	}
	G[0][0] = *n - 1
	G[*n-1][1] = 0
	for i := range G {
		x := rander.Intn(numLinks)
		G[i][0], G[i][x] = G[i][x], G[i][0]
		x = rander.Intn(numLinks)
		G[i][1], G[i][x] = G[i][x], G[i][1]
	}
	log.Printf("Graph Completed: %s", time.Since(now))
	return G
}

var t = template.Must(template.New("page").Parse(`<!DOCTYPE html>
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
    <p>TODO: generate random body content</p>
	</body>
</html>`))

type server struct {
	g             graph
	hit, mis, bad uint64
}

func (s *server) page(w http.ResponseWriter, r *http.Request) {
	// now := time.Now()
	u, err := strconv.Atoi(r.URL.Path[6:]) // 6 = len("/page/")
	if err != nil {
		http.NotFound(w, r)
		atomic.AddUint64(&s.bad, 1)
		return
	}
	if u > *n || u < 0 {
		s.redirect(w, r)
		return
	}

	links := make([]string, len(s.g[u]))
	for i, n := range s.g[u] {
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
	atomic.AddUint64(&s.hit, 1)
}

func (s *server) redirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/page/"+pad(rand.Intn(*n)), http.StatusTemporaryRedirect)
	atomic.AddUint64(&s.mis, 1)
}

func (s *server) logger() {
	var total, val uint64
	for {
		if val = atomic.LoadUint64(&s.hit); val != 0 {
			total += val
			log.Printf("QPS: %d; Total: %d\n", val, total)
			atomic.StoreUint64(&s.hit, 0)
		}
		time.Sleep(time.Second)
	}
}

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())
	s := &server{g: genGraph(*l)}
	http.HandleFunc("/page/", s.page)
	http.HandleFunc("/", s.redirect)
	go s.logger()
	http.ListenAndServe(":"+strconv.Itoa(*port), nil)
}
