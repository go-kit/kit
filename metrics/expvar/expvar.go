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
//
// Histogram observation count is exported as a counter called name + "_count"
package expvar

import (
	"expvar"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/codahale/hdrhistogram"

	"github.com/adrianco/kit/metrics"
)

type counter struct {
	v *expvar.Int
}

func (c counter) String() string { return fmt.Sprintf("%v", c.v) }

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

func (g gauge) String() string { return fmt.Sprintf("%v", g.v) }

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
	count  metrics.Counter
	gauges map[int]metrics.Gauge
}

// Name returns the name of the histogram
func (h *histogram) Name() string {
	return h.name
}

// Print out nonzero bars of the histogram as a csv with normalized probability and a crude bar graph
func (h histogram) String() string {
	var total float64
	d := h.hist.Merge().Distribution()
	for _, b := range d {
		total += float64(b.Count)
	}
	f := "%8v,%8v,%8v,%7v, %v\n"
	bs := "####################################################################################################"
	flbs := float64(len(bs))
	s := fmt.Sprintf(f, "From", "To", "Count", "Prob", "Bar")
	for _, b := range d {
		if b.Count > 0 {
			p := float64(b.Count) / total
			s += fmt.Sprintf(f, b.From, b.To, b.Count, fmt.Sprintf("%0.4f", p), "|"+bs[:int(p*flbs)])
		}
	}
	return fmt.Sprintf("name: %v\ncount: %v\ngauges: %v\n%v\n", h.name, h.count, h.gauges, s)
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
	h.count = NewCounter(name + "_count")
	go h.rotateLoop(1 * time.Minute)
	return h
}

func (h *histogram) With(metrics.Field) metrics.Histogram { return h }

func (h *histogram) Observe(value int64) {
	h.mu.Lock()
	err := h.hist.Current.RecordValue(value)
	h.count.Add(1)
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
