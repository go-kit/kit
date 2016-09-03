package pcp

import (
	"github.com/go-kit/kit/metrics"
	"github.com/performancecopilot/speed"
)

// Counter implements metrics.Counter via a single dimensional speed.Counter
// for now, see https://github.com/performancecopilot/speed/issues/32
type Counter struct {
	c speed.Counter
}

// NewCounter creates a new Counter
func NewCounter(name string) *Counter {
	c, err := speed.NewPCPCounter(0, name)
	if err != nil {
		panic(err)
	}

	registry.AddMetric(c)
	if err != nil {
		panic(err)
	}

	return &Counter{c}
}

// With is a no-op.
func (c *Counter) With(labelValues ...string) metrics.Counter { return c }

// Add implements Counter.
// speed Counters only take int64
// if it is important, instead use speed.SingletonMetric with DoubleType, CounterSemantics and OneUnit
// but that will mean this will need a mutex, to be safe
func (c *Counter) Add(delta float64) { c.c.Inc(int64(delta)) }

////////////////////////////////////////////////////////////////////////////////////////

// Gauge implements metrics.Gauge
// also singleton for now, for same reasons as Counter
type Gauge struct {
	g speed.Gauge
}

// NewGauge creates a new Gauge
func NewGauge(name string) *Gauge {
	g, err := speed.NewPCPGauge(0, name)
	if err != nil {
		panic(err)
	}

	err = registry.AddMetric(g)
	if err != nil {
		panic(err)
	}

	return &Gauge{g}
}

// With is a no-op.
func (g *Gauge) With(labelValues ...string) metrics.Gauge { return g }

// Set sets the value of the gauge
func (g *Gauge) Set(value float64) { g.g.Set(value) }

// Add adds a value to the gauge
func (g *Gauge) Add(value float64) { g.g.Inc(value) }

////////////////////////////////////////////////////////////////////////////////////////

// Histogram wraps a PCP Histogram
type Histogram struct {
	h speed.Histogram
}

// NewHistogram creates a new Histogram
// minimum observeable value is 0
// maximum observeable value is 3600000000
func NewHistogram(name string) *Histogram {
	h, err := speed.NewPCPHistogram(name, 0, 3600000000, 5)
	if err != nil {
		panic(err)
	}

	err = registry.AddMetric(h)
	if err != nil {
		panic(err)
	}

	return &Histogram{h}
}

// With is a no-op.
func (h *Histogram) With(labelValues ...string) metrics.Histogram { return h }

// Observe observes a value
//
// this converts float64 value to int64, as the Histogram in speed
// is backed using codahale/hdrhistogram, which only observes int64 values
func (h *Histogram) Observe(value float64) { h.h.MustRecord(int64(value)) }

// Mean returns the mean of the values observed so far
func (h *Histogram) Mean() float64 { return h.h.Mean() }

// Percentile returns a percentile between 0 and 100
func (h *Histogram) Percentile(p float64) int64 { return h.h.Percentile(p) }

////////////////////////////////////////////////////////////////////////////

var registry *speed.PCPRegistry
var client *speed.PCPClient

func init() {
	registry = speed.NewPCPRegistry()
}

// StartReporting starts reporting currently registered metrics to the PCP backend
func StartReporting(appname string) {
	if client != nil {
		panic("reporting is already enabled")
	}

	var err error
	client, err = speed.NewPCPClientWithRegistry(appname, registry)
	if err != nil {
		panic(err)
	}

	client.MustStart()
}

// StopReporting stops all reporting
func StopReporting() {
	if client == nil {
		panic("reporting is not active")
	}

	client.MustStop()
	client = nil
}
