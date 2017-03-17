package stats

import (
	"fmt"
	"log"
	"os"
	"sort"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

type msg struct {
	Type string          `json:"typ"`
	Msg  map[string]stat `json:"msg"`
}

type stat struct {
	Total uint64 `json:"tot"`
	More  uint64 `json:"new"`
}

// Stats is a general statistics object
type Stats struct {
	boring bool // if we have already broadcasted no updates, don't spam
	totals map[string]uint64
	nums   map[string]uint64
	nuMu   sync.Mutex
	name   string
	ws     *websocket.Conn
}

// NewStats constructs a new stats object
func NewStats(name, kind string) *Stats {
	// TODO(bign8): backoff and jitter retries - https://www.awsarchitectureblog.com/2015/03/backoff.html
	ws, err := websocket.Dial("ws://"+os.Getenv("ADMIN")+"/ws/"+kind, "", "http://"+name)
	if err != nil {
		log.Println("Websocket cannot connect", err)
		ws = nil
	}
	return &Stats{
		totals: make(map[string]uint64),
		nums:   make(map[string]uint64),
		name:   kind + "." + name,
		ws:     ws,
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

	//	if we have nothing to say, only update the first time
	if all == 0 {
		if !s.boring {
			if err := websocket.JSON.Send(s.ws, msg{Type: "stat", Msg: clone}); err != nil {
				panic(err) // TODO: handle errors better
			}
			s.boring = true
		}
		return ""
	}
	s.boring = false

	// Sort and print result (TODO: remove sort when only sending to admin)
	sort.Strings(keys)
	batch := make([]string, 0, len(keys))
	for _, key := range keys {
		batch = append(batch, fmt.Sprintf("%s(%d)%d", key, clone[key].More, clone[key].Total))
	}
	if s.ws != nil {
		if err := websocket.JSON.Send(s.ws, msg{Type: "stat", Msg: clone}); err != nil {
			panic(err) // TODO: handle errors better
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
