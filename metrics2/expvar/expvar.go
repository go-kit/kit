// Package expvar provides an expvar backend for metrics.
//
// All metric names support a primitive form of templates, to support Go kit's
// With/keyvals mechanism of establishing dimensionality. The behavior is best
// illustrated with an example.
//
//    p := NewProvider(...)
//    c := p.NewIntCounter("foo_{x}_{y}_bar")
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
	ints       map[string]*expvar.Int
	floats     map[string]*expvar.Float
	histograms map[string]*histogram.Histogram // demuxed to per-quantile gauges
}

// NewProvider returns a new, empty provider.
func NewProvider() *Provider {
	return &Provider{
		ints:       map[string]*expvar.Int{},
		floats:     map[string]*expvar.Float{},
		histograms: map[string]*histogram.Histogram{},
	}
}

// NewCounter is an alias for NewFloatCounter.
func (p *Provider) NewCounter(id metrics.Identifier) (metrics.Counter, error) {
	return p.NewFloatCounter(id), nil
}

// NewIntCounter returns a Counter whose values are truncated and exposed as
// expvar.Int.
//
// Only the NameTemplate field from the identifier is used. It can include
// template interpolation to support With; see package documentation for
// details.
func (p *Provider) NewIntCounter(id metrics.Identifier) *IntCounter {
	return &IntCounter{
		parent:  p,
		name:    id.NameTemplate,
		keyvals: keyval.MakeWith(template.ExtractKeysFrom(id.NameTemplate)),
	}
}

// NewFloatCounter returns a Counter whose values are exposed as an
// expvar.Float.
//
// Only the NameTemplate field from the identifier is used. It can include
// template interpolation to support With; see package documentation for
// details.
func (p *Provider) NewFloatCounter(id metrics.Identifier) *FloatCounter {
	return &FloatCounter{
		parent:  p,
		name:    id.NameTemplate,
		keyvals: keyval.MakeWith(template.ExtractKeysFrom(id.NameTemplate)),
	}
}

// NewGauge is an alias for NewFloatGauge.
func (p *Provider) NewGauge(id metrics.Identifier) (metrics.Gauge, error) {
	return p.NewFloatGauge(id), nil
}

// NewIntGauge returns a Gauge whose values are truncated and exposed as
// expvar.Int.
//
// Only the NameTemplate field from the identifier is used. It can include
// template interpolation to support With; see package documentation for
// details.
func (p *Provider) NewIntGauge(id metrics.Identifier) *IntGauge {
	return &IntGauge{
		parent:  p,
		name:    id.NameTemplate,
		keyvals: keyval.MakeWith(template.ExtractKeysFrom(id.NameTemplate)),
	}
}

// NewFloatGauge returns a Gauge whose values are exposed as an expvar.Float.
//
// Only the NameTemplate field from the identifier is used. It can include
// template interpolation to support With; see package documentation for
// details.
func (p *Provider) NewFloatGauge(id metrics.Identifier) *FloatGauge {
	return &FloatGauge{
		parent:  p,
		name:    id.NameTemplate,
		keyvals: keyval.MakeWith(template.ExtractKeysFrom(id.NameTemplate)),
	}
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

func (p *Provider) int(name string) *expvar.Int {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	if _, ok := p.ints[name]; !ok {
		p.ints[name] = expvar.NewInt(name)
	}
	return p.ints[name]
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

// IntCounter is a Counter whose values are truncated to integers and exposed as
// expvar.Int. IntCounters must be constructed via the Provider; the zero value
// of an IntCounter is not useful.
type IntCounter struct {
	parent  *Provider
	name    string
	keyvals map[string]string
}

// With implements Counter.
func (c *IntCounter) With(keyvals ...string) metrics.Counter {
	return &IntCounter{
		parent:  c.parent,
		name:    c.name,
		keyvals: keyval.Merge(c.keyvals, keyvals...),
	}
}

// Add inplements Counter.
func (c *IntCounter) Add(delta float64) {
	name := template.Render(c.name, c.keyvals)
	c.parent.int(name).Add(int64(delta))
}

// FloatCounter is a Counter whose values are exposed as expvar.Float.
// FloatCounters must be constructed via the Provider; the zero value of a
// FloatCounter is not useful.
type FloatCounter struct {
	parent  *Provider
	name    string
	keyvals map[string]string
}

// With implements Counter.
func (c *FloatCounter) With(keyvals ...string) metrics.Counter {
	return &FloatCounter{
		parent:  c.parent,
		name:    c.name,
		keyvals: keyval.Merge(c.keyvals, keyvals...),
	}
}

// Add implements Counter.
func (c *FloatCounter) Add(delta float64) {
	name := template.Render(c.name, c.keyvals)
	c.parent.float(name).Add(delta)
}

// IntGauge is a Gauge whose values are truncated and exposed as expvar.Int.
// IntGauges must be constructed via the Provider; the zero value of an IntGauge
// is not useful.
type IntGauge struct {
	parent  *Provider
	name    string
	keyvals map[string]string
}

// With implements Gauge.
func (g *IntGauge) With(keyvals ...string) metrics.Gauge {
	return &IntGauge{
		parent:  g.parent,
		name:    g.name,
		keyvals: keyval.Merge(g.keyvals, keyvals...),
	}
}

// Set implements Gauge.
func (g *IntGauge) Set(value float64) {
	name := template.Render(g.name, g.keyvals)
	g.parent.int(name).Set(int64(value))
}

// Add implements Gauge.
func (g *IntGauge) Add(delta float64) {
	name := template.Render(g.name, g.keyvals)
	g.parent.int(name).Add(int64(delta))
}

// FloatGauge is a Gauge whose values are exposed as expvar.Float. FloatGauges
// must be constructed via the Provider; the zero value of a FloatGauge is not
// useful.
type FloatGauge struct {
	parent  *Provider
	name    string
	keyvals map[string]string
}

// With implements Gauge.
func (g *FloatGauge) With(keyvals ...string) metrics.Gauge {
	return &FloatGauge{
		parent:  g.parent,
		name:    g.name,
		keyvals: keyval.Merge(g.keyvals, keyvals...),
	}
}

// Set implements Gauge.
func (g *FloatGauge) Set(value float64) {
	name := template.Render(g.name, g.keyvals)
	g.parent.float(name).Set(value)
}

// Add implements Gauge.
func (g *FloatGauge) Add(delta float64) {
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
