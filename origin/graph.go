package main

import (
	"log"
	"math/rand"
	"time"
)

type graph [][]int

func (g graph) Size() int { return len(g) }

func genGraph(size, numLinks int) graph {
	rander := rand.New(rand.NewSource(*seed))
	if numLinks > size {
		numLinks = size - 1
	}
	log.Print(host + ": Starting Graph Creation.")
	now := time.Now()
	G := make(graph, size)
	for i := range G {
		G[i] = make([]int, numLinks)
		G[i][0] = i - 1
		G[i][1] = i + 1
		for j := 2; j < numLinks; j++ {
			G[i][j] = rander.Intn(size)
		}
	}
	G[0][0] = size - 1
	G[size-1][1] = 0
	for i := range G {
		x := rander.Intn(numLinks)
		G[i][0], G[i][x] = G[i][x], G[i][0]
		x = rander.Intn(numLinks)
		G[i][1], G[i][x] = G[i][x], G[i][1]
	}
	log.Printf(host+": Graph Completed: %s", time.Since(now))
	return G
}
