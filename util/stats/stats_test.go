package stats

import (
	"sync"
	"sync/atomic"
	"testing"
)

type oldModel struct {
	totals map[string]uint64
	counts map[string]uint64
	cmutex sync.Mutex
}

func (old *oldModel) Inc(name string) {
	old.cmutex.Lock()
	old.counts[name]++
	old.cmutex.Unlock()
}

func (old *oldModel) Val(name string) uint64 {
	old.cmutex.Lock()
	cnt := old.counts[name]
	old.cmutex.Unlock()
	return cnt
}

type newModel struct {
	bins map[string]newSubModel
	mutx sync.Mutex
}

type newSubModel struct {
	total  uint64
	counts uint64
}

func (nm *newModel) Inc(name string) {
	nm.mutx.Lock()
	obj := nm.bins[name]
	nm.mutx.Unlock()
	atomic.AddUint64(&obj.counts, 1)
	// obj.counts++
}

func (nm *newModel) Val(name string) uint64 {
	nm.mutx.Lock()
	val := nm.bins[name].counts
	nm.mutx.Unlock()
	return val
}

type threeModel struct {
	bins map[string]*atomic.Value
	mu   sync.Mutex
}

func (tm *threeModel) Inc(name string) {
	tm.mu.Lock()
	obj := tm.bins[name]
	tm.mu.Unlock()
	if obj == nil {
		var alloc atomic.Value
		alloc.Store(uint64(1))
		tm.bins[name] = &alloc
		return
	}
	val, ok := obj.Load().(uint64)
	if !ok {
		obj.Store(uint64(1))
	} else {
		obj.Store(val + 1)
	}
}

func (tm *threeModel) Val(name string) uint64 {
	tm.mu.Lock()
	obj := tm.bins[name]
	tm.mu.Unlock()
	num, ok := obj.Load().(uint64)
	if !ok {
		return 0
	}
	return num
}

type Subject interface {
	Inc(string)
	Val(string) uint64
}

func runBench(b *testing.B, subj Subject) {
	var wg sync.WaitGroup
	wg.Add(b.N)
	for i := 0; i < b.N; i++ {
		go func() {
			subj.Inc("num")
			wg.Done()
		}()
	}
	wg.Wait()
	if num := subj.Val("num"); int(num) != b.N {
		b.Fatalf("Inavlid Count: %d != %d", b.N, num)
	}
}

func BenchmarkStatsOne(b *testing.B) {
	s := &oldModel{
		totals: make(map[string]uint64),
		counts: make(map[string]uint64),
	}
	runBench(b, s)
}

// func BenchmarkStatsTwo(b *testing.B) {
// 	runBench(b, &newModel{
// 		bins: make(map[string]newSubModel),
// 	})
// }

// func BenchmarkStatsThree(b *testing.B) {
// 	runBench(b, &threeModel{
// 		bins: make(map[string]*atomic.Value),
// 	})
// }
