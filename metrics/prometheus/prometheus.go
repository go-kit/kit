// Package prometheus implements a Prometheus backend for package metrics.
package prometheus

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/peterbourgon/gokit/metrics"
)

// Prometheus has strong opinions about the dimensionality of fields. Users
// must predeclare every field key they intend to use. On every observation,
// fields with keys that haven't been predeclared will be silently dropped,
// and predeclared field keys without values will receive the value
// PrometheusLabelValueUnknown.
var PrometheusLabelValueUnknown = "unknown"

type prometheusCounter struct {
	*prometheus.CounterVec
	Pairs map[string]string
}

// NewCounter returns a new Counter backed by a Prometheus metric. The counter
// is automatically registered via prometheus.Register.
func NewCounter(namespace, subsystem, name, help string, fieldKeys []string) metrics.Counter {
	return NewCounterWithLabels(namespace, subsystem, name, help, fieldKeys, prometheus.Labels{})
}

// NewCounterWithLabels is the same as NewCounter, but attaches a set of const
// label pairs to the metric.
func NewCounterWithLabels(namespace, subsystem, name, help string, fieldKeys []string, constLabels prometheus.Labels) metrics.Counter {
	m := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        name,
			Help:        help,
			ConstLabels: constLabels,
		},
		fieldKeys,
	)
	prometheus.MustRegister(m)

	p := map[string]string{}
	for _, fieldName := range fieldKeys {
		p[fieldName] = PrometheusLabelValueUnknown
	}

	return prometheusCounter{
		CounterVec: m,
		Pairs:      p,
	}
}

func (c prometheusCounter) With(f metrics.Field) metrics.Counter {
	return prometheusCounter{
		CounterVec: c.CounterVec,
		Pairs:      merge(c.Pairs, f),
	}
}

func (c prometheusCounter) Add(delta uint64) {
	c.CounterVec.With(prometheus.Labels(c.Pairs)).Add(float64(delta))
}

type prometheusGauge struct {
	*prometheus.GaugeVec
	Pairs map[string]string
}

// NewGauge returns a new Gauge backed by a Prometheus metric. The gauge is
// automatically registered via prometheus.Register.
func NewGauge(namespace, subsystem, name, help string, fieldKeys []string) metrics.Gauge {
	return NewGaugeWithLabels(namespace, subsystem, name, help, fieldKeys, prometheus.Labels{})
}

// NewGaugeWithLabels is the same as NewGauge, but attaches a set of const
// label pairs to the metric.
func NewGaugeWithLabels(namespace, subsystem, name, help string, fieldKeys []string, constLabels prometheus.Labels) metrics.Gauge {
	m := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        name,
			Help:        help,
			ConstLabels: constLabels,
		},
		fieldKeys,
	)
	prometheus.MustRegister(m)

	return prometheusGauge{
		GaugeVec: m,
		Pairs:    pairsFrom(fieldKeys),
	}
}

func (g prometheusGauge) With(f metrics.Field) metrics.Gauge {
	return prometheusGauge{
		GaugeVec: g.GaugeVec,
		Pairs:    merge(g.Pairs, f),
	}
}

func (g prometheusGauge) Set(value int64) {
	g.GaugeVec.With(prometheus.Labels(g.Pairs)).Set(float64(value))
}

func (g prometheusGauge) Add(delta int64) {
	g.GaugeVec.With(prometheus.Labels(g.Pairs)).Add(float64(delta))
}

type prometheusGaugeFloat struct {
	*prometheus.GaugeVec
	Pairs map[string]string
}

// NewGaugeFloat returns a new GaugeFloat backed by a Prometheus metric. The gauge is
// automatically registered via prometheus.Register.
func NewGaugeFloat(namespace, subsystem, name, help string, fieldKeys []string) metrics.GaugeFloat {
	return NewGaugeFloatWithLabels(namespace, subsystem, name, help, fieldKeys, prometheus.Labels{})
}

// NewGaugeWithLabels is the same as NewGauge, but attaches a set of const
// label pairs to the metric.
func NewGaugeFloatWithLabels(namespace, subsystem, name, help string, fieldKeys []string, constLabels prometheus.Labels) metrics.GaugeFloat {
	m := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        name,
			Help:        help,
			ConstLabels: constLabels,
		},
		fieldKeys,
	)
	prometheus.MustRegister(m)

	return prometheusGaugeFloat{
		GaugeVec: m,
		Pairs:    pairsFrom(fieldKeys),
	}
}

func (g prometheusGaugeFloat) With(f metrics.Field) metrics.GaugeFloat {
	return prometheusGaugeFloat{
		GaugeVec: g.GaugeVec,
		Pairs:    merge(g.Pairs, f),
	}
}

func (g prometheusGaugeFloat) Set(value float64) {
	g.GaugeVec.With(prometheus.Labels(g.Pairs)).Set(value)
}

func (g prometheusGaugeFloat) Add(delta float64) {
	g.GaugeVec.With(prometheus.Labels(g.Pairs)).Add(delta)
}

type prometheusHistogram struct {
	*prometheus.SummaryVec
	Pairs map[string]string
}

// NewHistogram returns a new Histogram backed by a Prometheus summary. It
// uses a 10-second max age for bucketing. The histogram is automatically
// registered via prometheus.Register.
func NewHistogram(namespace, subsystem, name, help string, fieldKeys []string) metrics.Histogram {
	return NewHistogramWithLabels(namespace, subsystem, name, help, fieldKeys, prometheus.Labels{})
}

// NewHistogramWithLabels is the same as NewHistogram, but attaches a set of
// const label pairs to the metric.
func NewHistogramWithLabels(namespace, subsystem, name, help string, fieldKeys []string, constLabels prometheus.Labels) metrics.Histogram {
	m := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        name,
			Help:        help,
			ConstLabels: constLabels,
			MaxAge:      10 * time.Second,
		},
		fieldKeys,
	)
	prometheus.MustRegister(m)

	return prometheusHistogram{
		SummaryVec: m,
		Pairs:      pairsFrom(fieldKeys),
	}
}

func (h prometheusHistogram) With(f metrics.Field) metrics.Histogram {
	return prometheusHistogram{
		SummaryVec: h.SummaryVec,
		Pairs:      merge(h.Pairs, f),
	}
}

func (h prometheusHistogram) Observe(value int64) {
	h.SummaryVec.With(prometheus.Labels(h.Pairs)).Observe(float64(value))
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
