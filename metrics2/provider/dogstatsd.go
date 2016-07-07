package provider

import (
	"github.com/go-kit/kit/metrics2"
	"github.com/go-kit/kit/metrics2/dogstatsd"
)

// Dogstatsd
type dogstatsdProvider struct {
	d    *dogstatsd.Dogstatsd
	stop func()
}

// NewDogstatsdProvider wraps the given Dogstatsd object and stop func and
// returns a Provider that produces Dogstatsd metrics.
func NewDogstatsdProvider(d *dogstatsd.Dogstatsd, stop func()) Provider {
	return &dogstatsdProvider{
		d:    d,
		stop: stop,
	}
}

// NewCounter implements Provider.
func (p *dogstatsdProvider) NewCounter(name string) metrics.Counter {
	return p.d.NewCounter(name)
}

// NewGauge implements Provider.
func (p *dogstatsdProvider) NewGauge(name string) metrics.Gauge {
	return p.d.NewGauge(name)
}

// NewHistogram implements Provider, returning a new Dogstatsd Histogram with a
// sample rate of 1.0. Buckets are ignored.
func (p *dogstatsdProvider) NewHistogram(name string, _ int) metrics.Histogram {
	return p.d.NewHistogram(name, 1.0)
}

// Stop implements Provider, invoking the stop function passed at construction.
func (p *dogstatsdProvider) Stop() {
	p.stop()
}
