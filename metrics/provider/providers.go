package provider

import (
	"errors"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	kitexp "github.com/go-kit/kit/metrics/expvar"
	"github.com/go-kit/kit/metrics/graphite"
	kitprom "github.com/go-kit/kit/metrics/prometheus"
	"github.com/go-kit/kit/metrics/statsd"
)

// Provider represents a union set of constructors and lifecycle management
// functions for each supported metrics backend. It should be used by those who
// need to easily swap out implementations, e.g. dynamically, or at a single
// point in an intermediating framework.
type Provider interface {
	NewCounter(name, help string) metrics.Counter
	NewHistogram(name, help string, min, max int64, sigfigs int, quantiles ...int) (metrics.Histogram, error)
	NewGauge(name, help string) metrics.Gauge
	Stop()
}

// NewGraphiteProvider will return a Provider implementation that is a simple
// wrapper around a graphite.Emitter. All metric names will be prefixed with the
// given value and data will be emitted once every interval. If no network value
// is given, it will default to "udp".
func NewGraphiteProvider(network, address, prefix string, interval time.Duration, logger log.Logger) (Provider, error) {
	if network == "" {
		network = "udp"
	}
	if address == "" {
		return nil, errors.New("address is required")
	}
	return graphiteProvider{
		e: graphite.NewEmitter(network, address, prefix, interval, logger),
	}, nil
}

type graphiteProvider struct {
	e *graphite.Emitter
}

var _ Provider = graphiteProvider{}

// NewCounter implements Provider. Help is ignored.
func (p graphiteProvider) NewCounter(name, _ string) metrics.Counter {
	return p.e.NewCounter(name)
}

// NewHistogram implements Provider. Help is ignored.
func (p graphiteProvider) NewHistogram(name, _ string, min, max int64, sigfigs int, quantiles ...int) (metrics.Histogram, error) {
	return p.e.NewHistogram(name, min, max, sigfigs, quantiles...)
}

// NewGauge implements Provider. Help is ignored.
func (p graphiteProvider) NewGauge(name, _ string) metrics.Gauge {
	return p.e.NewGauge(name)
}

// Stop implements Provider.
func (p graphiteProvider) Stop() {
	p.e.Stop()
}

// NewStatsdProvider will return a Provider implementation that is a simple
// wrapper around a statsd.Emitter. All metric names will be prefixed with the
// given value and data will be emitted once every interval or when the buffer
// has reached its max size. If no network value is given, it will default to
// "udp".
func NewStatsdProvider(network, address, prefix string, interval time.Duration, logger log.Logger) (Provider, error) {
	if network == "" {
		network = "udp"
	}
	if address == "" {
		return nil, errors.New("address is required")
	}
	return statsdProvider{
		e: statsd.NewEmitter(network, address, prefix, interval, logger),
	}, nil
}

type statsdProvider struct {
	e *statsd.Emitter
}

var _ Provider = statsdProvider{}

// NewCounter implements Provider. Help is ignored.
func (p statsdProvider) NewCounter(name, _ string) metrics.Counter {
	return p.e.NewCounter(name)
}

// NewHistogram implements Provider. Help is ignored.
func (p statsdProvider) NewHistogram(name, _ string, min, max int64, sigfigs int, quantiles ...int) (metrics.Histogram, error) {
	return p.e.NewHistogram(name), nil
}

// NewGauge implements Provider. Help is ignored.
func (p statsdProvider) NewGauge(name, _ string) metrics.Gauge {
	return p.e.NewGauge(name)
}

// Stop will call the underlying statsd.Emitter's Stop method.
func (p statsdProvider) Stop() {
	p.e.Stop()
}

// NewExpvarProvider is a very thin wrapper over the expvar package.
// If a prefix is provided, it will prefix all metric names.
func NewExpvarProvider(prefix string) Provider {
	return expvarProvider{prefix: prefix}
}

type expvarProvider struct {
	prefix string
}

var _ Provider = expvarProvider{}

// NewCounter implements Provider. Help is ignored.
func (p expvarProvider) NewCounter(name, _ string) metrics.Counter {
	return kitexp.NewCounter(p.prefix + name)
}

// NewHistogram implements Provider. Help is ignored.
func (p expvarProvider) NewHistogram(name, _ string, min, max int64, sigfigs int, quantiles ...int) (metrics.Histogram, error) {
	return kitexp.NewHistogram(p.prefix+name, min, max, sigfigs, quantiles...), nil
}

// NewGauge implements Provider. Help is ignored.
func (p expvarProvider) NewGauge(name, _ string) metrics.Gauge {
	return kitexp.NewGauge(p.prefix + name)
}

// Stop is a no-op.
func (expvarProvider) Stop() {}

type prometheusProvider struct {
	namespace string
	subsystem string
}

var _ Provider = prometheusProvider{}

// NewPrometheusProvider returns a Prometheus provider that uses the provided
// namespace and subsystem for all metrics. Help is not provided, which is very
// much against Prometheus guidelines. Therefore, avoid using the Prometheus
// provider unless it is absolutely necessary; prefer constructing Prometheus
// metrics directly.
func NewPrometheusProvider(namespace, subsystem string) Provider {
	return prometheusProvider{
		namespace: namespace,
		subsystem: subsystem,
	}
}

// NewCounter implements Provider.
func (p prometheusProvider) NewCounter(name, help string) metrics.Counter {
	return kitprom.NewCounter(prometheus.CounterOpts{
		Namespace: p.namespace,
		Subsystem: p.subsystem,
		Name:      name,
		Help:      help,
	}, nil)
}

// NewHistogram ignores all parameters except name and help.
func (p prometheusProvider) NewHistogram(name, help string, _, _ int64, _ int, _ ...int) (metrics.Histogram, error) {
	return kitprom.NewHistogram(prometheus.HistogramOpts{
		Namespace: p.namespace,
		Subsystem: p.subsystem,
		Name:      name,
		Help:      help,
	}, nil), nil
}

// NewGauge implements Provider.
func (p prometheusProvider) NewGauge(name, help string) metrics.Gauge {
	return kitprom.NewGauge(prometheus.GaugeOpts{
		Namespace: p.namespace,
		Subsystem: p.subsystem,
		Name:      name,
		Help:      help,
	}, nil)
}

// Stop is a no-op.
func (prometheusProvider) Stop() {}
