package push

import (
	"math/rand"

	"github.com/go-kit/kit/metrics2"
	"github.com/go-kit/kit/metrics2/internal/lv"
)

// Add captures a single counter add invocation.
type Add struct {
	Name       string
	SampleRate float64
	LVs        []string
	Delta      float64
}

// Counter is a forwarding implementation of the counter metric.
type Counter struct {
	name       string
	sampleRate float64
	lvs        lv.LabelValues
	c          chan<- Add
}

// NewCounter returns a counter that sends adds on the channel.
func NewCounter(name string, sampleRate float64, c chan<- Add) *Counter {
	return &Counter{name: name, sampleRate: sampleRate, c: c}
}

// Add forwards the delta to the remote. If sample rate is less than 1.0,
// it may be a no-op, depending on the result of rand.Float.
func (c Counter) Add(delta float64) {
	if c.sampleRate < 1.0 && rand.Float64() > c.sampleRate {
		return
	}
	c.c <- Add{Name: c.name, SampleRate: c.sampleRate, LVs: c.lvs, Delta: delta}
}

// With returns a new metric forwarding to the same destination with the
// provided label values appended to any existing label values.
func (c Counter) With(labelValues ...string) metrics.Counter {
	return &Counter{
		name:       c.name,
		sampleRate: c.sampleRate,
		lvs:        c.lvs.With(labelValues...),
		c:          c.c,
	}
}

// Set captures a single gauge set invocation.
type Set struct {
	Name  string
	LVs   []string
	Value float64
}

// Gauge is a forwarding implementation of the gauge metric.
type Gauge struct {
	name string
	lvs  lv.LabelValues
	c    chan<- Set
}

// NewGauge returns a Gauge that sends sets on the channel.
func NewGauge(name string, c chan<- Set) *Gauge {
	return &Gauge{name: name, c: c}
}

// Set forwards the delta to the remote.
func (g Gauge) Set(value float64) {
	g.c <- Set{Name: g.name, LVs: g.lvs, Value: value}
}

// With returns a new metric forwarding to the same destination with the
// provided label values appended to any existing label values.
func (g Gauge) With(labelValues ...string) metrics.Gauge {
	return &Gauge{
		name: g.name,
		lvs:  g.lvs.With(labelValues...),
		c:    g.c,
	}
}

// Obv captures a single histogram observe invocation.
type Obv struct {
	Name       string
	SampleRate float64
	LVs        []string
	Value      float64
}

// Histogram is a forwarding implementation of the histogram metric.
type Histogram struct {
	name       string
	sampleRate float64
	lvs        lv.LabelValues
	c          chan<- Obv
}

// NewHistogram returns a Histogram that sends observes on the channel. Sample rate
// is consulted at observation time.
func NewHistogram(name string, sampleRate float64, c chan<- Obv) *Histogram {
	return &Histogram{name: name, sampleRate: sampleRate, c: c}
}

// Observe forwards the value to the remote. If sample rate is less than 1.0,
// this call may be a no-op, depending on the result of of rand.Float.
func (h Histogram) Observe(value float64) {
	if h.sampleRate < 1.0 && rand.Float64() > h.sampleRate {
		return
	}
	h.c <- Obv{Name: h.name, SampleRate: h.sampleRate, LVs: h.lvs, Value: value}
}

// With returns a new metric forwarding to the same destination with the
// provided label values appended to any existing label values.
func (h Histogram) With(labelValues ...string) metrics.Histogram {
	return &Histogram{
		name:       h.name,
		sampleRate: h.sampleRate,
		lvs:        h.lvs.With(labelValues...),
		c:          h.c,
	}
}
