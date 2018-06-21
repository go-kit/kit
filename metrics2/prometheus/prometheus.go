// Package prometheus provides a Prometheus backend for metrics.
//
// Go kit's With/keyvals mechanism of establishing dimensionality maps directly
// to Prometheus' concept of labels. Prometheus labels must be predeclared when
// constructing a metric, so extra Go kit keyvals will be silently dropped, and
// unspecified Prometheus label keys will get a value of metrics.UnknownValue.
package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"

	metrics "github.com/go-kit/kit/metrics2"
	"github.com/go-kit/kit/metrics2/internal/keyval"
)

// Provider constructs and stores Prometheus metrics. Provider must be
// constructed via NewProvider; the zero value of a provider is not useful.
type Provider struct {
	// Registerer is used to register constructed metrics.
	// By default, prometheus.DefaultRegisterer is used.
	Registerer prometheus.Registerer
}

// NewProvider returns a new, empty provider.
func NewProvider() *Provider {
	return &Provider{
		Registerer: prometheus.DefaultRegisterer,
	}
}

// NewCounter constructs a prometheus.CounterVec, registers it via the
// Provider's configured Registerer, and returns a Counter wrapping it.
//
// The namespace, subsystem, name, help, and labels fields from the identifier
// are used. Labels (keys) must be completely specified at construction time,
// and take the default value of metrics.UnknownValue.
func (p *Provider) NewCounter(id metrics.Identifier) metrics.Counter {
	c := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: id.Namespace,
		Subsystem: id.Subsystem,
		Name:      id.Name,
		Help:      id.Help,
	}, id.Labels)
	p.Registerer.MustRegister(c)
	keyvals := map[string]string{}
	for _, label := range id.Labels {
		keyvals[label] = metrics.UnknownValue
	}
	return &counter{
		counter: c,
		keyvals: keyvals,
	}
}

// NewGauge constructs a prometheus.GaugeVec, registers it via the
// Provider's configured Registerer, and returns a Gauge wrapping it.
//
// The namespace, subsystem, name, help, and labels fields from the identifier
// are used. Labels (keys) must be completely specified at construction time,
// and take the default value of metrics.UnknownValue.
func (p *Provider) NewGauge(id metrics.Identifier) metrics.Gauge {
	g := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: id.Namespace,
		Subsystem: id.Subsystem,
		Name:      id.Name,
		Help:      id.Help,
	}, id.Labels)
	p.Registerer.MustRegister(g)
	keyvals := map[string]string{}
	for _, label := range id.Labels {
		keyvals[label] = metrics.UnknownValue
	}
	return &gauge{
		gauge:   g,
		keyvals: keyvals,
	}
}

// NewHistogram constructs a prometheus.HistogramVec, registers it via the
// Provider's configured Registerer, and returns a Histogram wrapping it.
//
// The namespace, subsystem, name, help, buckets, and labels fields from the
// identifier are used. Labels (keys) must be completely specified at
// construction time, and take the default value of metrics.UnknownValue.
func (p *Provider) NewHistogram(id metrics.Identifier) metrics.Histogram {
	h := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: id.Namespace,
		Subsystem: id.Subsystem,
		Name:      id.Name,
		Help:      id.Help,
		Buckets:   id.Buckets,
	}, id.Labels)
	p.Registerer.MustRegister(h)
	keyvals := map[string]string{}
	for _, label := range id.Labels {
		keyvals[label] = metrics.UnknownValue
	}
	return &histogram{
		histogram: h,
		keyvals:   keyvals,
	}
}

type counter struct {
	counter *prometheus.CounterVec
	keyvals map[string]string
}

func (c *counter) With(keyvals ...string) metrics.Counter {
	return &counter{
		counter: c.counter,
		keyvals: keyval.Merge(c.keyvals, keyvals...),
	}
}

func (c *counter) Add(value float64) {
	c.counter.With(prometheus.Labels(c.keyvals)).Add(value)
}

type gauge struct {
	gauge   *prometheus.GaugeVec
	keyvals map[string]string
}

func (g *gauge) With(keyvals ...string) metrics.Gauge {
	return &gauge{
		gauge:   g.gauge,
		keyvals: keyval.Merge(g.keyvals, keyvals...),
	}
}

func (g *gauge) Add(value float64) {
	g.gauge.With(prometheus.Labels(g.keyvals)).Add(value)
}

func (g *gauge) Set(value float64) {
	g.gauge.With(prometheus.Labels(g.keyvals)).Set(value)
}

type histogram struct {
	histogram *prometheus.HistogramVec
	keyvals   map[string]string
}

func (h *histogram) With(keyvals ...string) metrics.Histogram {
	return &histogram{
		histogram: h.histogram,
		keyvals:   keyval.Merge(h.keyvals, keyvals...),
	}
}

func (h *histogram) Observe(value float64) {
	h.histogram.With(prometheus.Labels(h.keyvals)).Observe(value)
}
