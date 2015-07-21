// Package expvar implements an expvar backend for package metrics.
//
// The current implementation ignores fields. In the future, it would be good
// to have an implementation that accepted a set of predeclared field names at
// construction time, and used field values to produce delimiter-separated
// bucket (key) names. That is,
//
//    c := NewFieldedCounter(..., "path", "status")
//    c.Add(1) // "myprefix_unknown_unknown" += 1
//    c2 := c.With("path", "foo").With("status": "200")
//    c2.Add(1) // "myprefix_foo_200" += 1
//
// It would also be possible to have an implementation that generated more
// sophisticated expvar.Values. For example, a Counter could be implemented as
// a map, representing a tree of key/value pairs whose leaves were the actual
// expvar.Ints.
package expvar

import (
	"expvar"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/codahale/hdrhistogram"

	"github.com/go-kit/kit/metrics"
)

type counter struct {
	v *expvar.Int
}

// NewCounter returns a new Counter backed by an expvar with the given name.
// Fields are ignored.
func NewCounter(name string) metrics.Counter {
	return &counter{expvar.NewInt(name)}
}

func (c *counter) With(metrics.Field) metrics.Counter { return c }
func (c *counter) Add(delta uint64)                   { c.v.Add(int64(delta)) }

type gauge struct {
	v *expvar.Float
}

// NewGauge returns a new Gauge backed by an expvar with the given name. It
// should be updated manually; for a callback-based approach, see
// PublishCallbackGauge. Fields are ignored.
func NewGauge(name string) metrics.Gauge {
	return &gauge{expvar.NewFloat(name)}
}

func (g *gauge) With(metrics.Field) metrics.Gauge { return g }

func (g *gauge) Add(delta float64) { g.v.Add(delta) }

func (g *gauge) Set(value float64) { g.v.Set(value) }

// PublishCallbackGauge publishes a Gauge as an expvar with the given name,
// whose value is determined at collect time by the passed callback function.
// The callback determines the value, and fields are ignored, so
// PublishCallbackGauge returns nothing.
func PublishCallbackGauge(name string, callback func() float64) {
	expvar.Publish(name, callbackGauge(callback))
}

type callbackGauge func() float64

func (g callbackGauge) String() string { return strconv.FormatFloat(g(), 'g', -1, 64) }

type histogram struct {
	mu   sync.Mutex
	hist *hdrhistogram.WindowedHistogram

	name   string
	gauges map[int]metrics.Gauge
}

// NewHistogram is taken from http://github.com/codahale/metrics. It returns a
// windowed HDR histogram which drops data older than five minutes.
//
// The histogram exposes metrics for each passed quantile as gauges. Quantiles
// should be integers in the range 1..99. The gauge names are assigned by
// using the passed name as a prefix and appending "_pNN" e.g. "_p50".
func NewHistogram(name string, minValue, maxValue int64, sigfigs int, quantiles ...int) metrics.Histogram {
	gauges := map[int]metrics.Gauge{}
	for _, quantile := range quantiles {
		if quantile <= 0 || quantile >= 100 {
			panic(fmt.Sprintf("invalid quantile %d", quantile))
		}
		gauges[quantile] = NewGauge(fmt.Sprintf("%s_p%02d", name, quantile))
	}
	h := &histogram{
		hist:   hdrhistogram.NewWindowed(5, minValue, maxValue, sigfigs),
		name:   name,
		gauges: gauges,
	}
	go h.rotateLoop(1 * time.Minute)
	return h
}

func (h *histogram) With(metrics.Field) metrics.Histogram { return h }

func (h *histogram) Observe(value int64) {
	h.mu.Lock()
	err := h.hist.Current.RecordValue(value)
	h.mu.Unlock()

	if err != nil {
		panic(err.Error())
	}

	for q, gauge := range h.gauges {
		gauge.Set(float64(h.hist.Current.ValueAtQuantile(float64(q))))
	}
}

func (h *histogram) rotateLoop(d time.Duration) {
	for range time.Tick(d) {
		h.mu.Lock()
		h.hist.Rotate()
		h.mu.Unlock()
	}
}
