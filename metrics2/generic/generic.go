// Package generic implements generic versions of each of the metric types. They
// can be embedded by other implementations, and converted to specific formats
// as necessary.
package generic

import (
	"math"
	"sync"
	"sync/atomic"

	"github.com/VividCortex/gohistogram"

	"github.com/go-kit/kit/metrics2"
)

// LabelValueUnknown is used as a label value when one is expected but not
// provided, typically due to user error.
const LabelValueUnknown = "unknown"

// Counter is an in-memory implementation of a Counter.
type Counter struct {
	sampleRate float64
	bits       uint64
	lvs        []string // immutable
}

// NewCounter returns a new, usable Counter.
func NewCounter() *Counter {
	return &Counter{}
}

// With implements Counter.
func (c *Counter) With(labelValues ...string) metrics.Counter {
	if len(labelValues)%2 != 0 {
		labelValues = append(labelValues, LabelValueUnknown)
	}
	return &Counter{
		bits: atomic.LoadUint64(&c.bits),
		lvs:  append(c.lvs, labelValues...),
	}
}

// Add implements Counter.
func (c *Counter) Add(delta float64) {
	for {
		var (
			old  = atomic.LoadUint64(&c.bits)
			newf = math.Float64frombits(old) + delta
			new  = math.Float64bits(newf)
		)
		if atomic.CompareAndSwapUint64(&c.bits, old, new) {
			break
		}
	}
}

// Value returns the current value of the counter.
func (c *Counter) Value() float64 {
	return math.Float64frombits(atomic.LoadUint64(&c.bits))
}

// ValueReset returns the current value of the counter, and resets it to zero.
// This is useful for metrics backends whose counter aggregations expect deltas,
// like statsd.
func (c *Counter) ValueReset() float64 {
	for {
		var (
			old  = atomic.LoadUint64(&c.bits)
			newf = 0.0
			new  = math.Float64bits(newf)
		)
		if atomic.CompareAndSwapUint64(&c.bits, old, new) {
			return math.Float64frombits(old)
		}
	}
}

// LabelValues returns the set of label values attached to the counter.
func (c *Counter) LabelValues() []string {
	return c.lvs
}

// Gauge is an in-memory implementation of a Gauge.
type Gauge struct {
	bits uint64
	lvs  []string // immutable
}

// NewGauge returns a new, usable Gauge.
func NewGauge() *Gauge {
	return &Gauge{}
}

// With implements Gauge.
func (c *Gauge) With(labelValues ...string) metrics.Gauge {
	if len(labelValues)%2 != 0 {
		labelValues = append(labelValues, LabelValueUnknown)
	}
	return &Gauge{
		bits: atomic.LoadUint64(&c.bits),
		lvs:  append(c.lvs, labelValues...),
	}
}

// Set implements Gauge.
func (c *Gauge) Set(value float64) {
	atomic.StoreUint64(&c.bits, math.Float64bits(value))
}

// Value returns the current value of the gauge.
func (c *Gauge) Value() float64 {
	return math.Float64frombits(atomic.LoadUint64(&c.bits))
}

// LabelValues returns the set of label values attached to the gauge.
func (c *Gauge) LabelValues() []string {
	return c.lvs
}

// Histogram is an in-memory implementation of a streaming histogram, based on
// VividCortex/gohistogram. It dynamically computes quantiles, so it's not
// suitable for aggregation.
type Histogram struct {
	lvs []string // immutable
	h   gohistogram.Histogram
}

// NewHistogram returns a numeric histogram based on VividCortex/gohistogram. A
// good default value for buckets is 50.
func NewHistogram(buckets int) *Histogram {
	return &Histogram{
		h: gohistogram.NewHistogram(buckets),
	}
}

// With implements Histogram.
func (h *Histogram) With(labelValues ...string) metrics.Histogram {
	if len(labelValues)%2 != 0 {
		labelValues = append(labelValues, LabelValueUnknown)
	}
	return &Histogram{
		lvs: append(h.lvs, labelValues...),
		h:   h.h,
	}
}

// Observe implements Histogram.
func (h *Histogram) Observe(value float64) {
	h.h.Add(value)
}

// Quantile returns the value of the quantile q, 0.0 < q < 1.0.
func (h *Histogram) Quantile(q float64) float64 {
	return h.h.Quantile(q)
}

// LabelValues returns the set of label values attached to the histogram.
func (h *Histogram) LabelValues() []string {
	return h.lvs
}

// SimpleHistogram is an in-memory implementation of a Histogram. It only tracks
// an approximate moving average, so is likely too naïve for many use cases.
type SimpleHistogram struct {
	mtx sync.RWMutex
	lvs []string
	avg float64
	n   uint64
}

// NewSimpleHistogram returns a SimpleHistogram, ready for observations.
func NewSimpleHistogram() *SimpleHistogram {
	return &SimpleHistogram{}
}

// With implements Histogram.
func (h *SimpleHistogram) With(labelValues ...string) metrics.Histogram {
	if len(labelValues)%2 != 0 {
		labelValues = append(labelValues, LabelValueUnknown)
	}
	return &SimpleHistogram{
		lvs: append(h.lvs, labelValues...),
		avg: h.avg,
		n:   h.n,
	}
}

// Observe implements Histogram.
func (h *SimpleHistogram) Observe(value float64) {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	h.n++
	h.avg -= h.avg / float64(h.n)
	h.avg += value / float64(h.n)
}

// ApproximateMovingAverage returns the approximate moving average of observations.
func (h *SimpleHistogram) ApproximateMovingAverage() float64 {
	h.mtx.RLock()
	h.mtx.RUnlock()
	return h.avg
}

// LabelValues returns the set of label values attached to the histogram.
func (h *SimpleHistogram) LabelValues() []string {
	return h.lvs
}
