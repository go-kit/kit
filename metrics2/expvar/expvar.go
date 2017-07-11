package expvar

import (
	"expvar"
	"sync"

	"github.com/go-kit/kit/metrics2"
	"github.com/go-kit/kit/metrics2/internal/histogram"
	"github.com/go-kit/kit/metrics2/internal/template"
)

// Provider acts as a store of expvar metrics.
//
type Provider struct {
	mtx        sync.Mutex
	ints       map[string]*expvar.Int
	floats     map[string]*expvar.Float
	histograms map[string]*histogram.Histogram //
}

// NewProvider returns a new, empty provider.
func NewProvider() *Provider {
	return &Provider{
		ints:       map[string]*expvar.Int{},
		floats:     map[string]*expvar.Float{},
		histograms: map[string]*histogram.Histogram{},
	}
}

// NewIntCounter returns a Counter whose values are truncated and exposed as
// expvar.Int. The name can include template interpolation to support With; see
// package documentation for details.
func (p *Provider) NewIntCounter(name string) *IntCounter {
	return &IntCounter{
		parent: p,
		name:   name,
		fields: map[string]string{},
	}
}

// NewFloatCounter returns a Counter whose values are exposed as an
// expvar.Float. The name can include template interpolation to support With;
// see package documentation for details.
func (p *Provider) NewFloatCounter(name string) *FloatCounter {
	return &FloatCounter{
		parent: p,
		name:   name,
		fields: map[string]string{},
	}
}

// NewIntGauge returns a Gauge whose values are truncated and exposed as
// expvar.Int. The name can include template interpolation to support With; see
// package documentation for details.
func (p *Provider) NewIntGauge(name string) *IntGauge {
	return &IntGauge{
		parent: p,
		name:   name,
		fields: map[string]string{},
	}
}

// NewFloatGauge returns a Gauge whose values are exposed as an expvar.Float.
// The name can include template interpolation to support With; see package
// documentation for details.
func (p *Provider) NewFloatGauge(name string) *FloatGauge {
	return &FloatGauge{
		parent: p,
		name:   name,
		fields: map[string]string{},
	}
}

// NewHistogram returns a Histogram whose observations are collected and exposed
// as 4 per-quantile expvar.Floats. The exposed quantiles are 50th, 90th, 95th,
// and 99th percentile, with names suffixed by _p50, _p90, _p95, and _p99,
// respectively. The name can include template interpolation to support With;
// see package documentation for details.
func (p *Provider) NewHistogram(name string) *Histogram {
	return &Histogram{
		parent: p,
		name:   name,
		fields: map[string]string{},
	}
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

	h, ok := p.histograms[name]
	if !ok {
		h = histogram.New()
		p.histograms[name] = h
	}
	h.Observe(value)

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
// expvar.Int. The zero value of an IntCounter is not useful; you must construct
// it via the Provider.
type IntCounter struct {
	parent *Provider
	name   string
	fields map[string]string
}

// With implements Counter.
func (c *IntCounter) With(keyvals ...string) metrics.Counter {
	return &IntCounter{
		parent: c.parent,
		name:   c.name,
		fields: appendFields(c.fields, keyvals...),
	}
}

// Add inplements Counter.
func (c *IntCounter) Add(delta float64) {
	name := template.Render(c.name, c.fields)
	c.parent.int(name).Add(int64(delta))
}

// FloatCounter is a Counter whose values are exposed as expvar.Float. The zero
// value of an FloatCounter is not useful; you must construct it via the
// Provider.
type FloatCounter struct {
	parent *Provider
	name   string
	fields map[string]string
}

// With implements Counter.
func (c *FloatCounter) With(keyvals ...string) metrics.Counter {
	return &FloatCounter{
		parent: c.parent,
		name:   c.name,
		fields: appendFields(c.fields, keyvals...),
	}
}

// Add implements Counter.
func (c *FloatCounter) Add(delta float64) {
	name := template.Render(c.name, c.fields)
	c.parent.float(name).Add(delta)
}

// IntGauge is a Gauge whose values are truncated and exposed as expvar.Int. The
// zero value of an IntGauge is not useful; you must construct it via the
// Provider.
type IntGauge struct {
	parent *Provider
	name   string
	fields map[string]string
}

// With implements Gauge.
func (g *IntGauge) With(keyvals ...string) metrics.Gauge {
	return &IntGauge{
		parent: g.parent,
		name:   g.name,
		fields: appendFields(g.fields, keyvals...),
	}
}

// Set implements Gauge.
func (g *IntGauge) Set(value float64) {
	name := template.Render(g.name, g.fields)
	g.parent.int(name).Set(int64(value))
}

// Add implements Gauge.
func (g *IntGauge) Add(delta float64) {
	name := template.Render(g.name, g.fields)
	g.parent.int(name).Add(int64(delta))
}

// FloatGauge is a Gauge whose values are exposed as expvar.Float. The zero
// value of a FloatGauge is not useful; you must construct it via the Provider.
type FloatGauge struct {
	parent *Provider
	name   string
	fields map[string]string
}

// With implements Gauge.
func (g *FloatGauge) With(keyvals ...string) metrics.Gauge {
	return &FloatGauge{
		parent: g.parent,
		name:   g.name,
		fields: appendFields(g.fields, keyvals...),
	}
}

// Set implements Gauge.
func (g *FloatGauge) Set(value float64) {
	name := template.Render(g.name, g.fields)
	g.parent.float(name).Set(value)
}

// Add implements Gauge.
func (g *FloatGauge) Add(delta float64) {
	name := template.Render(g.name, g.fields)
	g.parent.float(name).Add(delta)
}

// Histogram collects observations and exposes them as per-quantile Gauges. The
// zero value of a Histogram is not useful; you must construct it via that
// Provider.
type Histogram struct {
	parent *Provider
	name   string
	fields map[string]string
}

// With implements Histogram.
func (h *Histogram) With(keyvals ...string) metrics.Histogram {
	return &Histogram{
		parent: h.parent,
		name:   h.name,
		fields: appendFields(h.fields, keyvals...),
	}
}

// Observe implements Histogram.
func (h *Histogram) Observe(value float64) {
	name := template.Render(h.name, h.fields)
	h.parent.observe(name, value)
}

func appendFields(originalFields map[string]string, keyvals ...string) map[string]string {
	if len(keyvals)%2 != 0 {
		keyvals = append(keyvals, "unknown")
	}
	result := map[string]string{}
	for k, v := range originalFields {
		result[k] = v
	}
	for i := 0; i < len(keyvals); i += 2 {
		result[keyvals[i]] = keyvals[i+1]
	}
	return result
}
