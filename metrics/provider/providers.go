package provider

import (
	"errors"
	"net"
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
func NewGraphiteProvider(addr, net, prefix string, interval time.Duration, logger log.Logger) (Provider, error) {
	if addr == "" {
		return nil, errors.New("graphite server address is required")
	}
	if net == "" {
		net = "udp"
	}
	// nop logger for now :\
	e := graphite.NewEmitter(addr, net, prefix, interval, logger)
	return &graphiteProvider{Emitter: e}, nil
}

type graphiteProvider struct {
	*graphite.Emitter
}

// NewStatsdProvider will create a UDP connection for each metric
// with the given address. All metrics will use the given interval
// and, if a prefix is provided, it will be included in metric names
// with this format:
//    "prefix.name"
func NewStatsdProvider(addr, prefix string, interval time.Duration, logger log.Logger) (Provider, error) {
	return &statsdProvider{addr: addr, prefix: prefix, interval: interval}, nil
}

type statsdProvider struct {
	addr string

	interval time.Duration
	prefix   string

	logger log.Logger
}

func (s *statsdProvider) conn() (net.Conn, error) {
	return net.Dial("udp", s.addr)
}

func (s *statsdProvider) pref(name string) string {
	if len(s.prefix) > 0 {
		return s.prefix + "." + name
	}
	return name
}

func (s *statsdProvider) NewCounter(name string) metrics.Counter {
	conn, err := s.conn()
	if err != nil {
		s.logger.Log("during", "new counter", "err", err)
		return nil
	}
	return statsd.NewCounter(conn, s.pref(name), s.interval)
}

func (s *statsdProvider) NewHistogram(name string, min, max int64, sigfigs int, quantiles ...int) (metrics.Histogram, error) {
	conn, err := s.conn()
	if err != nil {
		return nil, err
	}
	return statsd.NewHistogram(conn, s.pref(name), s.interval), nil
}

func (s *statsdProvider) NewGauge(name string) metrics.Gauge {
	conn, err := s.conn()
	if err != nil {
		s.logger.Log("during", "new gauge", "err", err)
		return nil
	}
	return statsd.NewGauge(conn, s.pref(name), s.interval)
}

// Stop is a no-op  (should we try to close the UDP connections here?)
func (s *statsdProvider) Stop() {}

// NewExpvarProvider is a very thin wrapper over the expvar package.
// If a prefix is provided, it will be included in metric names with this
// format:
//    "prefix.name"
func NewExpvarProvider(prefix string) Provider {
	return &expvarProvider{prefix: prefix}
}

type expvarProvider struct {
	prefix string
}

func (e *expvarProvider) pref(name string) string {
	if len(e.prefix) > 0 {
		return e.prefix + "." + name
	}
	return name
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
