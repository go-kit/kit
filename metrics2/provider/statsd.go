package provider

import (
	"time"

	"github.com/go-kit/kit/metrics2"
	"github.com/go-kit/kit/metrics2/statsd"
)

type statsdProvider struct {
	s    *statsd.Statsd
	stop func()
}

// NewStatsdProvider wraps the given Statsd object and stop func and returns a
// Provider that produces Statsd metrics.
func NewStatsdProvider(s *statsd.Statsd, stop func()) Provider {
	return &statsdProvider{
		s:    s,
		stop: stop,
	}
}

// NewCounter implements Provider.
func (p *statsdProvider) NewCounter(name string) metrics.Counter {
	return p.s.NewCounter(name)
}

// NewGauge implements Provider.
func (p *statsdProvider) NewGauge(name string) metrics.Gauge {
	return p.s.NewGauge(name)
}

// NewHistogram implements Provider, returning a Histogram that accepts
// observations in seconds, and reports observations to Statsd in milliseconds.
// The sample rate is fixed at 1.0. The bucket parameter is ignored.
func (p *statsdProvider) NewHistogram(name string, _ int) metrics.Histogram {
	return p.s.MustNewHistogram(name, time.Second, time.Millisecond, 1.0)
}

// Stop implements Provider, invoking the stop function passed at construction.
func (p *statsdProvider) Stop() {
	p.stop()
}
