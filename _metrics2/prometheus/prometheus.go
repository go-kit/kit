// Package prometheus provides Prometheus implementations for metrics.
// Individual metrics are mapped to their Prometheus counterparts, and
// (depending on the constructor used) may be automatically registered in the
// global Prometheus metrics registry.
package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/go-kit/kit/metrics2"
	"github.com/go-kit/kit/metrics2/generic"
)

// Counter implements Counter, via a Prometheus CounterVec.
type Counter struct {
	cv *prometheus.CounterVec
	lv []string
}

// NewCounterFrom constructs and registers a Prometheus CounterVec,
// and returns a usable Counter object.
func NewCounterFrom(opts prometheus.CounterOpts, labelNames []string) *Counter {
	cv := prometheus.NewCounterVec(opts, labelNames)
	prometheus.MustRegister(cv)
	return NewCounter(cv)
}

// NewCounter wraps the CounterVec and returns a usable Counter object.
func NewCounter(cv *prometheus.CounterVec) *Counter {
	return &Counter{
		cv: cv,
		lv: []string{},
	}
}

// With implements Counter.
func (c *Counter) With(labelValues ...string) metrics.Counter {
	if len(labelValues)%2 != 0 {
		labelValues = append(labelValues, generic.LabelValueUnknown)
	}
	return &Counter{
		cv: c.cv,
		lv: append(c.lv, labelValues...),
	}
}

// Add implements Counter.
func (c *Counter) Add(delta float64) {
	c.cv.WithLabelValues(c.lv...).Add(delta)
}

// Gauge implements Gauge, via a Prometheus GaugeVec.
type Gauge struct {
	gv *prometheus.GaugeVec
	lv []string
}

// NewGaugeFrom construts and registers a Prometheus GaugeVec,
// and returns a usable Gauge object.
func NewGaugeFrom(opts prometheus.GaugeOpts, labelNames []string) *Gauge {
	gv := prometheus.NewGaugeVec(opts, labelNames)
	prometheus.MustRegister(gv)
	return NewGauge(gv)
}

// NewGauge wraps the GaugeVec and returns a usable Gauge object.
func NewGauge(gv *prometheus.GaugeVec) *Gauge {
	return &Gauge{
		gv: gv,
		lv: []string{},
	}
}

// With implements Gauge.
func (g *Gauge) With(labelValues ...string) metrics.Gauge {
	if len(labelValues)%2 != 0 {
		labelValues = append(labelValues, generic.LabelValueUnknown)
	}
	return &Gauge{
		gv: g.gv,
		lv: append(g.lv, labelValues...),
	}
}

// Set implements Gauge.
func (g *Gauge) Set(value float64) {
	g.gv.WithLabelValues(g.lv...).Set(value)
}

// Add is supported by Prometheus GaugeVecs.
func (g *Gauge) Add(delta float64) {
	g.gv.WithLabelValues(g.lv...).Add(delta)
}

// Summary implements Histogram, via a Prometheus SummaryVec. The difference
// between a Summary and a Histogram is that Summaries don't require predefined
// quantile buckets, but cannot be statistically aggregated.
type Summary struct {
	sv *prometheus.SummaryVec
	lv []string
}

// NewSummaryFrom constructs and registers a Prometheus SummaryVec,
// and returns a usable Summary object.
func NewSummaryFrom(opts prometheus.SummaryOpts, labelNames []string) *Summary {
	sv := prometheus.NewSummaryVec(opts, labelNames)
	prometheus.MustRegister(sv)
	return NewSummary(sv)
}

// NewSummary wraps the SummaryVec and returns a usable Summary object.
func NewSummary(sv *prometheus.SummaryVec) *Summary {
	return &Summary{
		sv: sv,
		lv: []string{},
	}
}

// With implements Histogram.
func (s *Summary) With(labelValues ...string) metrics.Histogram {
	if len(labelValues)%2 != 0 {
		labelValues = append(labelValues, generic.LabelValueUnknown)
	}
	return &Summary{
		sv: s.sv,
		lv: append(s.lv, labelValues...),
	}
}

// Observe implements Histogram.
func (s *Summary) Observe(value float64) {
	s.sv.WithLabelValues(s.lv...).Observe(value)
}

// Histogram implements Histogram via a Prometheus HistogramVec. The difference
// between a Histogram and a Summary is that Histograms require predefined
// quantile buckets, and can be statistically aggregated.
type Histogram struct {
	hv *prometheus.HistogramVec
	lv []string
}

// NewHistogramFrom constructs and registers a Prometheus HistogramVec,
// and returns a usable Histogram object.
func NewHistogramFrom(opts prometheus.HistogramOpts, labelNames []string) *Histogram {
	hv := prometheus.NewHistogramVec(opts, labelNames)
	prometheus.MustRegister(hv)
	return NewHistogram(hv)
}

// NewHistogram wraps the HistogramVec and returns a usable Histogram object.
func NewHistogram(hv *prometheus.HistogramVec) *Histogram {
	return &Histogram{
		hv: hv,
		lv: []string{},
	}
}

// With implements Histogram.
func (h *Histogram) With(labelValues ...string) metrics.Histogram {
	if len(labelValues)%2 != 0 {
		labelValues = append(labelValues, generic.LabelValueUnknown)
	}
	return &Histogram{
		hv: h.hv,
		lv: append(h.lv, labelValues...),
	}
}

// Observe implements Histogram.
func (h *Histogram) Observe(value float64) {
	h.hv.WithLabelValues(h.lv...).Observe(value)
}
