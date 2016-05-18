package provider

import (
	"errors"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	kitexp "github.com/go-kit/kit/metrics/expvar"
	"github.com/go-kit/kit/metrics/graphite"
	kitprom "github.com/go-kit/kit/metrics/prometheus"
	"github.com/go-kit/kit/metrics/statsd"
	"github.com/prometheus/client_golang/prometheus"
)

// Provider represents a union set of constructors and lifecycle management
// functions for each supported metrics backend. It should be used by those
// who need to easily swap out implementations, e.g. dynamically, or at a
// single point in an intermediating framework.
type Provider interface {
	NewCounter(name string) metrics.Counter
	NewHistogram(name string, min, max int64, sigfigs int, quantiles ...int) (metrics.Histogram, error)
	NewGauge(name string) metrics.Gauge

	Stop()
}

// NewGraphiteProvider will return a Provider implementation that is a simple
// wrapper around a graphite.Emitter. All metrics names will get prefixed
// with the given value and data will be emitted once every interval.
// If no network value is given, it will get defaulted to "udp".
func NewGraphiteProvider(addr, network, prefix string, interval time.Duration, logger log.Logger) (Provider, error) {
	if addr == "" {
		return nil, errors.New("graphite server address is required")
	}
	if network == "" {
		network = "udp"
	}
	e := graphite.NewEmitter(addr, network, prefix, interval, logger)
	return &graphiteProvider{Emitter: e}, nil
}

type graphiteProvider struct {
	*graphite.Emitter
}

// NewStatsdProvider will return a Provider implementation that is a simple
// wrapper around a statsd.Emitter. All metrics names will get prefixed
// with the given value and data will be emitted once every interval
// or when the buffer has reached its max size.
// If no network value is given, it will get defaulted to "udp".
func NewStatsdProvider(addr, network, prefix string, interval time.Duration, logger log.Logger) (Provider, error) {
	if addr == "" {
		return nil, errors.New("statsd server address is required")
	}
	if network == "" {
		network = "udp"
	}
	e := statsd.NewEmitter(addr, network, prefix, interval, logger)
	return &statsdProvider{e: e}, nil
}

type statsdProvider struct {
	e *statsd.Emitter
}

func (s *statsdProvider) NewCounter(name string) metrics.Counter {
	return s.e.NewCounter(name)
}

func (s *statsdProvider) NewHistogram(name string, min, max int64, sigfigs int, quantiles ...int) (metrics.Histogram, error) {
	return s.e.NewHistogram(name), nil
}

func (s *statsdProvider) NewGauge(name string) metrics.Gauge {
	return s.e.NewGauge(name)
}

// Stop will call the underlying statsd.Emitter's Stop method.
func (s *statsdProvider) Stop() {
	s.e.Stop()
}

// NewExpvarProvider is a very thin wrapper over the expvar package.
// If a prefix is provided, it will prefix in metric names.
func NewExpvarProvider(prefix string) Provider {
	return &expvarProvider{prefix: prefix}
}

type expvarProvider struct {
	prefix string
}

func (e *expvarProvider) pref(name string) string {
	return e.prefix + name
}

func (e *expvarProvider) NewCounter(name string) metrics.Counter {
	return kitexp.NewCounter(e.pref(name))
}

func (e *expvarProvider) NewHistogram(name string, min, max int64, sigfigs int, quantiles ...int) (metrics.Histogram, error) {
	return kitexp.NewHistogram(e.pref(name), min, max, sigfigs, quantiles...), nil
}

func (e *expvarProvider) NewGauge(name string) metrics.Gauge {
	return kitexp.NewGauge(e.pref(name))
}

// Stop is a no-op.
func (e *expvarProvider) Stop() {}

type prometheusProvider struct {
	ns string
}

// NewPrometheusProvider will use the given namespace
// for all metrics' Opts.
func NewPrometheusProvider(namespace string) Provider {
	return &prometheusProvider{ns: namespace}
}

func (p *prometheusProvider) NewCounter(name string) metrics.Counter {
	opts := prometheus.CounterOpts{
		Namespace: p.ns,
		Name:      name,
	}
	return kitprom.NewCounter(opts, nil)
}

// NewHistogram ignores all NewHistogram parameters but `name`.
func (p *prometheusProvider) NewHistogram(name string, min, max int64, sigfigs int, quantiles ...int) (metrics.Histogram, error) {
	opts := prometheus.HistogramOpts{
		Namespace: p.ns,
		Name:      name,
	}
	return kitprom.NewHistogram(opts, nil), nil
}

func (p *prometheusProvider) NewGauge(name string) metrics.Gauge {
	opts := prometheus.GaugeOpts{
		Namespace: p.ns,
		Name:      name,
	}
	return kitprom.NewGauge(opts, nil)
}

// Stop is a no-op
func (p *prometheusProvider) Stop() {}
