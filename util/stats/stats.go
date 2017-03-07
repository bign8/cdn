package stats

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	redis "gopkg.in/redis.v3"
)

type stat struct {
	Total uint64
	More  uint64
}

// Stats is a general statistics object
type Stats struct {
	totals map[string]uint64
	nums   map[string]uint64
	nuMu   sync.Mutex
	name   string
	conn   *redis.Client
}

// NewStats constructs a new stats object
func NewStats(name string, conn *redis.Client) *Stats {
	return &Stats{
		totals: make(map[string]uint64),
		nums:   make(map[string]uint64),
		name:   name,
		conn:   conn,
	}
}

// Inc adds a particular value
func (s *Stats) Inc(name string) {
	s.nuMu.Lock()
	s.nums[name]++
	s.nuMu.Unlock()
}

func (s *Stats) String() string {
	s.nuMu.Lock()
	var all uint64
	keys := make([]string, 0, len(s.nums))
	clone := make(map[string]stat, len(s.nums))
	for key, value := range s.nums {
		keys = append(keys, key)
		all += value
		s.totals[key] += value
		clone[key] = stat{
			More:  value,
			Total: s.totals[key],
		}
		s.nums[key] = 0
	}
	s.nuMu.Unlock()
	if all == 0 {
		return ""
	}
	sort.Strings(keys)
	batch := make([]string, 0, len(keys))
	for _, key := range keys {
		batch = append(batch, fmt.Sprintf("%s(%d)%d", key, clone[key].More, clone[key].Total))
	}
	if s.conn != nil {
		bits, err := json.Marshal(clone)
		if err != nil {
			panic(err)
		}
		if err = s.conn.Set(s.name, string(bits), time.Second).Err(); err != nil {
			panic(err)
		}
	}
	return fmt.Sprintf("all:%d; VPS(new)total: %s", all, batch) // Value per second
}

// Report stores data every interval with name
func (s *Stats) Report(interval time.Duration) {
	for c := time.Tick(interval); ; <-c {
		if line := s.String(); line != "" {
			log.Printf("%s: %s", s.name, line)
		}
	}
}
