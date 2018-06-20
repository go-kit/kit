// Package expvar provides an expvar backend for metrics.
//
// All metric names support a primitive form of templates, to support Go kit's
// With/keyvals mechanism of establishing dimensionality. The behavior is best
// illustrated with an example.
//
//    p := NewProvider(...)
//    c := p.NewCounter("foo_{x}_{y}_bar")
//    c.Add(1)                          // foo_unknown_unknown_bar += 1
//    c.With("x", "hello").Add(2)       // foo_hello_unknown_bar += 2
//    c.With("x", "1", "y", "2").Add(4) // foo_1_2_bar += 4
//    c.With("quux", "bing").Add(8)     // foo_unknown_unknown_bar += 8
//
package expvar

import (
	"expvar"
	"sync"

	"github.com/go-kit/kit/metrics2"
	"github.com/go-kit/kit/metrics2/internal/histogram"
	"github.com/go-kit/kit/metrics2/internal/keyval"
	"github.com/go-kit/kit/metrics2/internal/template"
)

// Provider constructs and stores expvar metrics.
type Provider struct {
	mtx        sync.Mutex
	floats     map[string]*expvar.Float
	histograms map[string]*histogram.Histogram // demuxed to per-quantile gauges
}

// NewProvider returns a new, empty provider.
func NewProvider() *Provider {
	return &Provider{
		floats:     map[string]*expvar.Float{},
		histograms: map[string]*histogram.Histogram{},
	}
}

// NewCounter returns a Counter whose values are exposed as an expvar.Float.
//
// Only the NameTemplate field from the identifier is used. It can include
// template interpolation to support With; see package documentation for
// details.
func (p *Provider) NewCounter(id metrics.Identifier) (metrics.Counter, error) {
	return &Counter{
		parent:  p,
		name:    id.NameTemplate,
		keyvals: keyval.MakeWith(template.ExtractKeysFrom(id.NameTemplate)),
	}, nil
}

// NewGauge returns a Gauge whose values are exposed as an expvar.Float.
//
// Only the NameTemplate field from the identifier is used. It can include
// template interpolation to support With; see package documentation for
// details.
func (p *Provider) NewGauge(id metrics.Identifier) (metrics.Gauge, error) {
	return &Gauge{
		parent:  p,
		name:    id.NameTemplate,
		keyvals: keyval.MakeWith(template.ExtractKeysFrom(id.NameTemplate)),
	}, nil
}

// NewHistogram returns a Histogram whose observations are collected and exposed
// as 4 per-quantile expvar.Floats. The exposed quantiles are 50th, 90th, 95th,
// and 99th percentile, with names suffixed by _p50, _p90, _p95, and _p99,
// respectively.
//
// Only the NameTemplate field from the identifier is used. It can include
// template interpolation to support With; see package documentation for
// details.
func (p *Provider) NewHistogram(id metrics.Identifier) (metrics.Histogram, error) {
	return &Histogram{
		parent:  p,
		name:    id.NameTemplate,
		keyvals: keyval.MakeWith(template.ExtractKeysFrom(id.NameTemplate)),
	}, nil
}

func (p *Provider) float(name string) *expvar.Float {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	if _, ok := p.floats[name]; !ok {
		p.floats[name] = expvar.NewFloat(name)
	}
	return p.floats[name]
}

func (p *Provider) observe(name string, value float64) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	// Observe the value in the histogram.
	h, ok := p.histograms[name]
	if !ok {
		h = histogram.New()
		p.histograms[name] = h
	}
	h.Observe(value)

	// Demux the histogram to per-quantile gauges.
	for _, pair := range []struct {
		suffix   string
		quantile float64
	}{
		{"_p50", 0.50},
		{"_p90", 0.90},
		{"_p95", 0.95},
		{"_p99", 0.99},
	} {
		fullname := name + pair.suffix
		f, ok := p.floats[fullname]
		if !ok {
			f = expvar.NewFloat(fullname)
			p.floats[fullname] = f
		}
		f.Set(h.Quantile(pair.quantile))
	}
}

// Counter exposes values as as expvar.Float. Counters must be constructed via
// the Provider; the zero value of a Counter is not useful.
type Counter struct {
	parent  *Provider
	name    string
	keyvals map[string]string
}

// With implements Counter.
func (c *Counter) With(keyvals ...string) metrics.Counter {
	return &Counter{
		parent:  c.parent,
		name:    c.name,
		keyvals: keyval.Merge(c.keyvals, keyvals...),
	}
}

// Add implements Counter.
func (c *Counter) Add(delta float64) {
	name := template.Render(c.name, c.keyvals)
	c.parent.float(name).Add(delta)
}

// Gauge exposes values as expvar.Float. Gauges must be constructed via the
// Provider; the zero value of a Gauge is not useful.
type Gauge struct {
	parent  *Provider
	name    string
	keyvals map[string]string
}

// With implements Gauge.
func (g *Gauge) With(keyvals ...string) metrics.Gauge {
	return &Gauge{
		parent:  g.parent,
		name:    g.name,
		keyvals: keyval.Merge(g.keyvals, keyvals...),
	}
}

// Set implements Gauge.
func (g *Gauge) Set(value float64) {
	name := template.Render(g.name, g.keyvals)
	g.parent.float(name).Set(value)
}

// Add implements Gauge.
func (g *Gauge) Add(delta float64) {
	name := template.Render(g.name, g.keyvals)
	g.parent.float(name).Add(delta)
}

// Histogram collects observations and exposes them as per-quantile Gauges.
// Histograms must be constructed via the Provider; the zero value of a
// Histogram is not useful.
type Histogram struct {
	parent  *Provider
	name    string
	keyvals map[string]string
}

// With implements Histogram.
func (h *Histogram) With(keyvals ...string) metrics.Histogram {
	return &Histogram{
		parent:  h.parent,
		name:    h.name,
		keyvals: keyval.Merge(h.keyvals, keyvals...),
	}
}

// Observe implements Histogram.
func (h *Histogram) Observe(value float64) {
	name := template.Render(h.name, h.keyvals)
	h.parent.observe(name, value)
}
