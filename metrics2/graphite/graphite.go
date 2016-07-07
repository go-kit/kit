// Package graphite provides a Graphite backend for metrics. Metrics are emitted
// with each observation in the plaintext protocol. See
// http://graphite.readthedocs.io/en/latest/feeding-carbon.html#the-plaintext-protocol
// for more information.
//
// Graphite does not have a native understanding of metric parameterization, so
// label values are aggregated but not reported. Use distinct metrics for each
// unique combination of label values.
package graphite

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics2/generic"
	"github.com/go-kit/kit/util/conn"
)

// Graphite is a store for metrics that will be reported to a Graphite server.
// Create a Graphite object, use it to create metrics objects, and pass those
// objects as dependencies to the components that will use them.
type Graphite struct {
	mtx        sync.RWMutex
	prefix     string
	counters   map[string]*generic.Counter
	gauges     map[string]*generic.Gauge
	histograms map[string]*generic.Histogram
	logger     log.Logger
}

// New creates a Statsd object that flushes all metrics in the Graphite
// plaintext format every flushInterval to the network and address. Use the
// returned stop function to terminate the flushing goroutine.
func New(prefix string, network, address string, flushInterval time.Duration, logger log.Logger) (res *Graphite, stop func()) {
	s := NewRaw(prefix, logger)
	manager := conn.NewDefaultManager(network, address, logger)
	ticker := time.NewTicker(flushInterval)
	go s.FlushTo(manager, ticker)
	return s, ticker.Stop
}

// NewRaw returns a Graphite object capable of allocating individual metrics.
// All metrics will share the given prefix in their path. All metrics can be
// snapshotted, and their values and statistical summaries written to a writer,
// via the WriteTo method.
func NewRaw(prefix string, logger log.Logger) *Graphite {
	return &Graphite{
		prefix:     prefix,
		counters:   map[string]*generic.Counter{},
		gauges:     map[string]*generic.Gauge{},
		histograms: map[string]*generic.Histogram{},
		logger:     logger,
	}
}

// NewCounter allocates and returns a counter with the given name.
func (g *Graphite) NewCounter(name string) *generic.Counter {
	g.mtx.Lock()
	defer g.mtx.Unlock()
	c := generic.NewCounter()
	g.counters[g.prefix+name] = c
	return c
}

// NewGauge allocates and returns a gauge with the given name.
func (g *Graphite) NewGauge(name string) *generic.Gauge {
	g.mtx.Lock()
	defer g.mtx.Unlock()
	ga := generic.NewGauge()
	g.gauges[g.prefix+name] = ga
	return ga
}

// NewHistogram allocates and returns a histogram with the given name and bucket
// count. 50 is a good default number of buckets. Histograms report their 50th,
// 90th, 95th, and 99th quantiles in distinct metrics with the .p50, .p90, .p95,
// and .p99 suffixes, respectively.
func (g *Graphite) NewHistogram(name string, buckets int) *generic.Histogram {
	g.mtx.Lock()
	defer g.mtx.Unlock()
	h := generic.NewHistogram(buckets)
	g.histograms[g.prefix+name] = h
	return h
}

// FlushTo invokes WriteTo to the writer every time the ticker fires. FlushTo
// blocks until the ticker is stopped. Most users won't need to call this method
// directly, and should prefer to use the New constructor.
func (g *Graphite) FlushTo(w io.Writer, ticker *time.Ticker) {
	for range ticker.C {
		if _, err := g.WriteTo(w); err != nil {
			g.logger.Log("during", "flush", "err", err)
		}
	}
}

// WriteTo writes a snapshot of all of the allocated metrics to the writer in
// the Graphite plaintext format. Clients probably shouldn't invoke this method
// directly, and should prefer using FlushTo, or the New constructor.
func (g *Graphite) WriteTo(w io.Writer) (int64, error) {
	g.mtx.RLock()
	defer g.mtx.RUnlock()
	var (
		n     int
		err   error
		count int64
		now   = time.Now().Unix()
	)
	for path, c := range g.counters {
		n, err = fmt.Fprintf(w, "%s.count %f %d\n", path, c.Value(), now)
		if err != nil {
			return count, err
		}
		count += int64(n)
	}
	for path, ga := range g.gauges {
		n, err = fmt.Fprintf(w, "%s %f %d\n", path, ga.Value(), now)
		if err != nil {
			return count, err
		}
		count += int64(n)
	}
	for path, h := range g.histograms {
		n, err = fmt.Fprintf(w, "%s.p50 %f %d\n", path, h.Quantile(0.50), now)
		n, err = fmt.Fprintf(w, "%s.p90 %f %d\n", path, h.Quantile(0.90), now)
		n, err = fmt.Fprintf(w, "%s.p95 %f %d\n", path, h.Quantile(0.95), now)
		n, err = fmt.Fprintf(w, "%s.p99 %f %d\n", path, h.Quantile(0.99), now)
		if err != nil {
			return count, err
		}
		count += int64(n)
	}
	return count, nil
}
