// Package discard implements a backend for package metrics that succeeds
// without doing anything.
package discard

import "github.com/go-kit/kit/metrics"

type counter struct {
	name string
}

// NewCounter returns a Counter that does nothing.
func NewCounter(name string) metrics.Counter { return &counter{name} }

func (c *counter) Name() string                       { return c.name }
func (c *counter) With(metrics.Field) metrics.Counter { return c }
func (c *counter) Add(delta uint64)                   {}

type gauge struct {
	name string
}

// NewGauge returns a Gauge that does nothing.
func NewGauge(name string) metrics.Gauge { return &gauge{name} }

func (g *gauge) Name() string                     { return g.name }
func (g *gauge) With(metrics.Field) metrics.Gauge { return g }
func (g *gauge) Set(value float64)                {}
func (g *gauge) Add(delta float64)                {}
func (g *gauge) Get() float64                     { return 0 }

type histogram struct {
	name string
}

// NewHistogram returns a Histogram that does nothing.
func NewHistogram(name string) metrics.Histogram { return &histogram{name} }

func (h *histogram) Name() string                         { return h.name }
func (h *histogram) With(metrics.Field) metrics.Histogram { return h }
func (h *histogram) Observe(value int64)                  {}
func (h *histogram) Distribution() ([]metrics.Bucket, []metrics.Quantile) {
	return []metrics.Bucket{}, []metrics.Quantile{}
}
