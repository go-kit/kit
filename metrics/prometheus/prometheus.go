// Package prometheus implements a Prometheus backend for package metrics.
package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/go-kit/kit/metrics"
)

// Prometheus has strong opinions about the dimensionality of fields. Users
// must predeclare every field key they intend to use. On every observation,
// fields with keys that haven't been predeclared will be silently dropped,
// and predeclared field keys without values will receive the value
// PrometheusLabelValueUnknown.
var PrometheusLabelValueUnknown = "unknown"

type prometheusCounter struct {
	*prometheus.CounterVec
	name  string
	Pairs map[string]string
}

// NewCounter returns a new Counter backed by a Prometheus metric. The counter
// is automatically registered via prometheus.Register.
func NewCounter(opts prometheus.CounterOpts, fieldKeys []string) metrics.Counter {
	m := prometheus.NewCounterVec(opts, fieldKeys)
	prometheus.MustRegister(m)
	p := map[string]string{}
	for _, fieldName := range fieldKeys {
		p[fieldName] = PrometheusLabelValueUnknown
	}
	return prometheusCounter{
		CounterVec: m,
		name:       opts.Name,
		Pairs:      p,
	}
}

func (c prometheusCounter) Name() string { return c.name }

func (c prometheusCounter) With(f metrics.Field) metrics.Counter {
	return prometheusCounter{
		CounterVec: c.CounterVec,
		name:       c.name,
		Pairs:      merge(c.Pairs, f),
	}
}

func (c prometheusCounter) Add(delta uint64) {
	c.CounterVec.With(prometheus.Labels(c.Pairs)).Add(float64(delta))
}

type prometheusGauge struct {
	*prometheus.GaugeVec
	name  string
	Pairs map[string]string
}

// NewGauge returns a new Gauge backed by a Prometheus metric. The gauge is
// automatically registered via prometheus.Register.
func NewGauge(opts prometheus.GaugeOpts, fieldKeys []string) metrics.Gauge {
	m := prometheus.NewGaugeVec(opts, fieldKeys)
	prometheus.MustRegister(m)
	return prometheusGauge{
		GaugeVec: m,
		name:     opts.Name,
		Pairs:    pairsFrom(fieldKeys),
	}
}

func (g prometheusGauge) Name() string { return g.name }

func (g prometheusGauge) With(f metrics.Field) metrics.Gauge {
	return prometheusGauge{
		GaugeVec: g.GaugeVec,
		name:     g.name,
		Pairs:    merge(g.Pairs, f),
	}
}

func (g prometheusGauge) Set(value float64) {
	g.GaugeVec.With(prometheus.Labels(g.Pairs)).Set(value)
}

func (g prometheusGauge) Add(delta float64) {
	g.GaugeVec.With(prometheus.Labels(g.Pairs)).Add(delta)
}

func (g prometheusGauge) Get() float64 {
	// TODO(pb): see https://github.com/prometheus/client_golang/issues/58
	return 0.0
}

// RegisterCallbackGauge registers a Gauge with Prometheus whose value is
// determined at collect time by the passed callback function. The callback
// determines the value, and fields are ignored, so RegisterCallbackGauge
// returns nothing.
func RegisterCallbackGauge(opts prometheus.GaugeOpts, callback func() float64) {
	prometheus.MustRegister(prometheus.NewGaugeFunc(opts, callback))
}

type prometheusSummary struct {
	*prometheus.SummaryVec
	name  string
	Pairs map[string]string
}

// NewSummary returns a new Histogram backed by a Prometheus summary. The
// histogram is automatically registered via prometheus.Register.
//
// For more information on Prometheus histograms and summaries, refer to
// http://prometheus.io/docs/practices/histograms.
func NewSummary(opts prometheus.SummaryOpts, fieldKeys []string) metrics.Histogram {
	m := prometheus.NewSummaryVec(opts, fieldKeys)
	prometheus.MustRegister(m)
	return prometheusSummary{
		SummaryVec: m,
		name:       opts.Name,
		Pairs:      pairsFrom(fieldKeys),
	}
}

func (s prometheusSummary) Name() string { return s.name }

func (s prometheusSummary) With(f metrics.Field) metrics.Histogram {
	return prometheusSummary{
		SummaryVec: s.SummaryVec,
		name:       s.name,
		Pairs:      merge(s.Pairs, f),
	}
}

func (s prometheusSummary) Observe(value int64) {
	s.SummaryVec.With(prometheus.Labels(s.Pairs)).Observe(float64(value))
}

func (s prometheusSummary) Distribution() ([]metrics.Bucket, []metrics.Quantile) {
	// TODO(pb): see https://github.com/prometheus/client_golang/issues/58
	return []metrics.Bucket{}, []metrics.Quantile{}
}

type prometheusHistogram struct {
	*prometheus.HistogramVec
	name  string
	Pairs map[string]string
}

// NewHistogram returns a new Histogram backed by a Prometheus Histogram. The
// histogram is automatically registered via prometheus.Register.
//
// For more information on Prometheus histograms and summaries, refer to
// http://prometheus.io/docs/practices/histograms.
func NewHistogram(opts prometheus.HistogramOpts, fieldKeys []string) metrics.Histogram {
	m := prometheus.NewHistogramVec(opts, fieldKeys)
	prometheus.MustRegister(m)
	return prometheusHistogram{
		HistogramVec: m,
		name:         opts.Name,
		Pairs:        pairsFrom(fieldKeys),
	}
}

func (h prometheusHistogram) Name() string { return h.name }

func (h prometheusHistogram) With(f metrics.Field) metrics.Histogram {
	return prometheusHistogram{
		HistogramVec: h.HistogramVec,
		name:         h.name,
		Pairs:        merge(h.Pairs, f),
	}
}

func (h prometheusHistogram) Observe(value int64) {
	h.HistogramVec.With(prometheus.Labels(h.Pairs)).Observe(float64(value))
}

func (h prometheusHistogram) Distribution() ([]metrics.Bucket, []metrics.Quantile) {
	// TODO(pb): see https://github.com/prometheus/client_golang/issues/58
	return []metrics.Bucket{}, []metrics.Quantile{}
}

func pairsFrom(fieldKeys []string) map[string]string {
	p := map[string]string{}
	for _, fieldName := range fieldKeys {
		p[fieldName] = PrometheusLabelValueUnknown
	}
	return p
}

func merge(orig map[string]string, f metrics.Field) map[string]string {
	if _, ok := orig[f.Key]; !ok {
		return orig
	}

	newPairs := make(map[string]string, len(orig))
	for k, v := range orig {
		newPairs[k] = v
	}

	newPairs[f.Key] = f.Value
	return newPairs
}
