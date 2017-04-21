package stats

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	metrics "github.com/rcrowley/go-metrics"
	"github.com/rcrowley/go-metrics/exp"
)

// New creates a new stats registry
func New(kind, name string, port int) Stats {
	registry := metrics.NewPrefixedRegistry(kind + ".")
	exp.Exp(registry)

	// NOTE: this is a HACK! (TODO: some awesome retries and stuff in the background)
	res, err := http.PostForm("http://"+os.Getenv("ADMIN")+"/hello", url.Values{
		"kind": {kind},
		"name": {name},
		"port": {strconv.Itoa(port)},
	})
	if err != nil {
		panic(err)
	}
	if res.StatusCode != http.StatusAccepted {
		log.Println("Admin not available!", res.Status)
	}
	return Stats{registry}.Sub(name)
}

// Stats is the same as Stats with more special sauce
type Stats struct {
	metrics.Registry
}

// Sub creaets a sub stats registry
func (s Stats) Sub(name string) Stats {
	if s.Registry == nil {
		return Stats{metrics.NewPrefixedRegistry(name + ".")} // For testing
	}
	return Stats{metrics.NewPrefixedChildRegistry(s.Registry, name+".")}
}

// Counter creates a counter
func (s Stats) Counter(name string) metrics.Counter {
	if s.Registry == nil {
		return &metrics.NilCounter{} // For testing
	}
	return metrics.GetOrRegisterCounter(name, s.Registry)
}

// Gauge creates a guage
func (s Stats) Gauge(name string) metrics.Gauge {
	if s.Registry == nil {
		return &metrics.NilGauge{} // For testing
	}
	return metrics.GetOrRegisterGauge(name, s.Registry)
}

// GaugeFloat64 creates a guage
func (s Stats) GaugeFloat64(name string) metrics.GaugeFloat64 {
	if s.Registry == nil {
		return &metrics.NilGaugeFloat64{} // For testing
	}
	return metrics.GetOrRegisterGaugeFloat64(name, s.Registry)
}

// Histogram creates a histogram
func (s Stats) Histogram(name string, sample metrics.Sample) metrics.Histogram {
	if s.Registry == nil {
		return &metrics.NilHistogram{} // For testing
	}
	return metrics.GetOrRegisterHistogram(name, s.Registry, sample)
}

// Meter creates a meter
func (s Stats) Meter(name string) metrics.Meter {
	if s.Registry == nil {
		return &metrics.NilMeter{} // For testing
	}
	return metrics.GetOrRegisterMeter(name, s.Registry)
}

// Timer creates a timer
func (s Stats) Timer(name string) metrics.Timer {
	if s.Registry == nil {
		return &metrics.NilTimer{} // For testing
	}
	return metrics.GetOrRegisterTimer(name, s.Registry)
}
