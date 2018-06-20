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

// Provider constructs and stores Prometheus metrics.
type Provider struct {
	registerer prometheus.Registerer
}

// NewProvider returns a new, empty provider.
func NewProvider(options ...ProviderOption) *Provider {
	p := &Provider{
		registerer: prometheus.DefaultRegisterer,
	}
	for _, option := range options {
		option(p)
	}
	return p
}

// ProviderOption modifies the behavior of the provider.
type ProviderOption func(*Provider)

// WithRegisterer changes the registry into which Prometheus metrics are
// registered. By default, the prometheus.DefaultRegisterer is used.
func WithRegisterer(r prometheus.Registerer) ProviderOption {
	return func(p *Provider) { p.registerer = r }
}

// NewCounter constructs a prometheus.CounterVec, registers it via the
// Provider's configured Registerer, and returns a Counter wrapping it.
//
// The namespace, subsystem, name, help, and labels fields from the identifier
// are used. Labels (keys) must be completely specified at construction time,
// and take the default value of metrics.UnknownValue.
func (p *Provider) NewCounter(id metrics.Identifier) (metrics.Counter, error) {
	c := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: id.Namespace,
		Subsystem: id.Subsystem,
		Name:      id.Name,
		Help:      id.Help,
	}, id.Labels)
	if err := p.registerer.Register(c); err != nil {
		return nil, err
	}
	keyvals := map[string]string{}
	for _, label := range id.Labels {
		keyvals[label] = metrics.UnknownValue
	}
	return &Counter{
		counter: c,
		keyvals: keyvals,
	}, nil
}

// NewGauge constructs a prometheus.GaugeVec, registers it via the
// Provider's configured Registerer, and returns a Gauge wrapping it.
//
// The namespace, subsystem, name, help, and labels fields from the identifier
// are used. Labels (keys) must be completely specified at construction time,
// and take the default value of metrics.UnknownValue.
func (p *Provider) NewGauge(id metrics.Identifier) (metrics.Gauge, error) {
	g := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: id.Namespace,
		Subsystem: id.Subsystem,
		Name:      id.Name,
		Help:      id.Help,
	}, id.Labels)
	if err := p.registerer.Register(g); err != nil {
		return nil, err
	}
	keyvals := map[string]string{}
	for _, label := range id.Labels {
		keyvals[label] = metrics.UnknownValue
	}
	return &Gauge{
		gauge:   g,
		keyvals: keyvals,
	}, nil
}

// NewHistogram constructs a prometheus.HistogramVec, registers it via the
// Provider's configured Registerer, and returns a Histogram wrapping it.
//
// The namespace, subsystem, name, help, and labels fields from the identifier
// are used. Labels (keys) must be completely specified at construction time,
// and take the default value of metrics.UnknownValue.
func (p *Provider) NewHistogram(id metrics.Identifier) (metrics.Histogram, error) {
	h := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: id.Namespace,
		Subsystem: id.Subsystem,
		Name:      id.Name,
		Help:      id.Help,
	}, id.Labels)
	if err := p.registerer.Register(h); err != nil {
		return nil, err
	}
	keyvals := map[string]string{}
	for _, label := range id.Labels {
		keyvals[label] = metrics.UnknownValue
	}
	return &Histogram{
		histogram: h,
		keyvals:   keyvals,
	}, nil
}

// Counter wraps a prometheus.CounterVec and implements metrics.Counter.
// Counters must be constructed via the Provider; the zero value of a Counter is
// not useful.
type Counter struct {
	counter *prometheus.CounterVec
	keyvals map[string]string
}

// With implements Counter. Keyvals whose keys haven't been predeclared as
// labels are silently dropped.
func (c *Counter) With(keyvals ...string) metrics.Counter {
	return &Counter{
		counter: c.counter,
		keyvals: keyval.Merge(c.keyvals, keyvals...),
	}
}

// Add implements Counter.
func (c *Counter) Add(value float64) {
	c.counter.With(prometheus.Labels(c.keyvals)).Add(value)
}

// Gauge wraps a prometheus.GaugeVec and implements metrics.Gauge. Gauges must
// be constructed via the Provider; the zero value of a Gauge is not useful.
type Gauge struct {
	gauge   *prometheus.GaugeVec
	keyvals map[string]string
}

// With implements Gauge. Keyvals whose keys haven't been predeclared as labels
// are silently dropped.
func (g *Gauge) With(keyvals ...string) metrics.Gauge {
	return &Gauge{
		gauge:   g.gauge,
		keyvals: keyval.Merge(g.keyvals, keyvals...),
	}
}

// Add implements Gauge.
func (g *Gauge) Add(value float64) {
	g.gauge.With(prometheus.Labels(g.keyvals)).Add(value)
}

// Set implements Gauge.
func (g *Gauge) Set(value float64) {
	g.gauge.With(prometheus.Labels(g.keyvals)).Set(value)
}

// Histogram wraps a prometheus.HistogramVec and implements metrics.Histogram.
// Histograms must be constructed via the Provider; the zero value of a
// Histogram is not useful.
type Histogram struct {
	histogram *prometheus.HistogramVec
	keyvals   map[string]string
}

// With implements Histogram. Keyvals whose keys haven't been predeclared as
// labels are silently dropped.
func (h *Histogram) With(keyvals ...string) metrics.Histogram {
	return &Histogram{
		histogram: h.histogram,
		keyvals:   keyval.Merge(h.keyvals, keyvals...),
	}
}

// Observe implements Histogram.
func (h *Histogram) Observe(value float64) {
	h.histogram.With(prometheus.Labels(h.keyvals)).Observe(value)
}
