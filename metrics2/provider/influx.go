package provider

import (
	"github.com/go-kit/kit/metrics2"
	"github.com/go-kit/kit/metrics2/influx"
)

type influxProvider struct {
	i    *influx.Influx
	stop func()
}

// NewInfluxProvider takes the given Influx object and stop func, and returns
// a Provider that produces Influx metrics.
func NewInfluxProvider(i *influx.Influx, stop func()) Provider {
	return &influxProvider{
		i:    i,
		stop: stop,
	}
}

// NewCounter implements Provider. Per-metric tags are not supported.
func (p *influxProvider) NewCounter(name string) metrics.Counter {
	return p.i.NewCounter(name, map[string]string{})
}

// NewGauge implements Provider. Per-metric tags are not supported.
func (p *influxProvider) NewGauge(name string) metrics.Gauge {
	return p.i.NewGauge(name, map[string]string{})
}

// NewHistogram implements Provider. Per-metric tags are not supported.
func (p *influxProvider) NewHistogram(name string, buckets int) metrics.Histogram {
	return p.i.NewHistogram(name, map[string]string{}, buckets)
}

// Stop implements Provider, invoking the stop function passed at construction.
func (p *influxProvider) Stop() {
	p.stop()
}
