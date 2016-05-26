// Package circonus provides a Circonus backend for package metrics.
//
// Users are responsible for calling the circonusgometrics.Start method
// themselves. Note that all Circonus metrics must have unique names, and are
// registered in a package-global registry. Circonus metrics also don't support
// fields, so all With methods are no-ops.
package circonus

import (
	"github.com/circonus-labs/circonus-gometrics"

	"github.com/go-kit/kit/metrics"
)

// NewCounter returns a counter backed by a Circonus counter with the given
// name. Due to the Circonus data model, fields are not supported.
func NewCounter(name string) metrics.Counter {
	return counter(name)
}

type counter circonusgometrics.Counter

// Name implements Counter.
func (c counter) Name() string {
	return string(c)
}

// With implements Counter, but is a no-op.
func (c counter) With(metrics.Field) metrics.Counter {
	return c
}

// Add implements Counter.
func (c counter) Add(delta uint64) {
	circonusgometrics.Counter(c).AddN(delta)
}

// NewGauge returns a gauge backed by a Circonus gauge with the given name. Due
// to the Circonus data model, fields are not supported. Also, Circonus gauges
// are defined as integers, so values are truncated.
func NewGauge(name string) metrics.Gauge {
	return gauge(name)
}

type gauge circonusgometrics.Gauge

// Name implements Gauge.
func (g gauge) Name() string {
	return string(g)
}

// With implements Gauge, but is a no-op.
func (g gauge) With(metrics.Field) metrics.Gauge {
	return g
}

// Set implements Gauge.
func (g gauge) Set(value float64) {
	circonusgometrics.Gauge(g).Set(int64(value))
}

// Add implements Gauge, but is a no-op, as Circonus gauges don't support
// incremental (delta) mutation.
func (g gauge) Add(float64) {
	return
}

// Get implements Gauge, but always returns zero, as there's no way to extract
// the current value from a Circonus gauge.
func (g gauge) Get() float64 {
	return 0.0
}

// NewHistogram returns a histogram backed by a Circonus histogram.
// Due to the Circonus data model, fields are not supported.
func NewHistogram(name string) metrics.Histogram {
	return histogram{
		h: circonusgometrics.NewHistogram(name),
	}
}

type histogram struct {
	h *circonusgometrics.Histogram
}

// Name implements Histogram.
func (h histogram) Name() string {
	return h.h.Name()
}

// With implements Histogram, but is a no-op.
func (h histogram) With(metrics.Field) metrics.Histogram {
	return h
}

// Observe implements Histogram. The value is converted to float64.
func (h histogram) Observe(value int64) {
	h.h.RecordValue(float64(value))
}

// Distribution implements Histogram, but is a no-op.
func (h histogram) Distribution() ([]metrics.Bucket, []metrics.Quantile) {
	return []metrics.Bucket{}, []metrics.Quantile{}
}
