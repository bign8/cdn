package main

import (
	"fmt"
	"log"
	"sort"
	"sync"
	"time"
)

type stats struct {
	totals map[string]uint64
	nums   map[string]uint64
	nuMu   sync.Mutex
}

func newStats() stats {
	return stats{
		totals: make(map[string]uint64),
		nums:   make(map[string]uint64),
	}
}

func (s *stats) inc(name string) {
	s.nuMu.Lock()
	s.nums[name]++
	s.nuMu.Unlock()
}

func (s *stats) String() string {
	s.nuMu.Lock()
	var all uint64
	keys := make([]string, 0, len(s.nums))
	clone := make(map[string]uint64, len(s.nums))
	for key, value := range s.nums {
		keys = append(keys, key)
		all += value
		clone[key] = value
		s.totals[key] += value
		s.nums[key] = 0
	}
	s.nuMu.Unlock()
	if all == 0 {
		return ""
	}
	sort.Strings(keys)
	batch := make([]string, 0, len(keys))
	for _, key := range keys {
		batch = append(batch, fmt.Sprintf("%s(%d)%d", key, clone[key], s.totals[key]))
	}
	return fmt.Sprintf("all:%d; VPS(new)total: %s", all, batch)
}

func (s *server) logger() {
	for c := time.Tick(*inter); ; <-c {
		if line := s.stat.String(); line != "" {
			log.Println(line)
		}
	}
}
