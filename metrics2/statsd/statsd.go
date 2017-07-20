package statsd

import (
	"fmt"
	"io"
	"sync"
	"time"
	"tmp/metrics"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics2/internal/keyval"
	"github.com/go-kit/kit/metrics2/internal/template"
	"github.com/go-kit/kit/util/conn"
)

// https://github.com/b/statsd_spec
type Provider struct {
	mtx         sync.RWMutex
	counters    map[string]float64
	gauges      map[string]float64
	timers      map[string][]float64
	histograms  map[string][]float64
	valueFormat string
	logger      log.Logger
}

type ProviderOption func(*Provider)

func WithFloatValues() ProviderOption {
	return func(p *Provider) { p.valueFormat = `%f` }
}

func NewProvider(logger log.Logger, options ...ProviderOption) *Provider {
	return &Provider{
		counters:    map[string]float64{},
		gauges:      map[string]float64{},
		timers:      map[string][]float64{},
		histograms:  map[string][]float64{},
		valueFormat: `%d`,
		logger:      logger,
	}
}

func (p *Provider) SendLoop(c <-chan time.Time, network, address string) {
	p.WriteLoop(c, conn.NewDefaultManager(network, address, p.logger))
}

func (p *Provider) WriteLoop(c <-chan time.Time, w io.Writer) {
	for range c {
		if err := p.WriteTo(w); err != nil {
			p.logger.Log("err", err)
		}
	}
}

func (p *Provider) WriteTo(w io.Writer) error {
	c, g, t, h := func() (c, g map[string]float64, t, h map[string][]float64) {
		p.mtx.Lock()
		defer p.mtx.Unlock()
		c, p.counters = p.counters, map[string]float64{}
		g, p.gauges = p.gauges, map[string]float64{}
		t, p.timers = p.timers, map[string][]float64{}
		h, p.histograms = p.histograms, map[string][]float64{}
		return
	}()
	for name, value := range c {
		if _, err := fmt.Fprintf(w, "%s:"+p.valueFormat+"|c\n", name, int(value)); err != nil {
			return err
		}
	}
	for name, value := range g {
		if _, err := fmt.Fprintf(w, "%s:"+p.valueFormat+"|g\n", name, int(value)); err != nil {
			return err
		}
	}
	for name, values := range t {
		for _, value := range values {
			if _, err := fmt.Fprintf(w, "%s:"+p.valueFormat+"|ms\n", name, int(value)); err != nil {
				return err
			}
		}
	}
	for name, values := range h {
		for _, value := range values {
			if _, err := fmt.Fprintf(w, "%s:"+p.valueFormat+"|h\n", name, int(value)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *Provider) NewCounter(name string) *Counter {
	return &Counter{
		parent:  p,
		name:    name,
		keyvals: map[string]string{},
	}
}

func (p *Provider) NewGauge(name string) *Gauge {
	return &Gauge{
		parent:  p,
		name:    name,
		keyvals: map[string]string{},
	}
}

func (p *Provider) NewTimer(name string) *Timer {
	return &Timer{
		parent:  p,
		name:    name,
		keyvals: map[string]string{},
	}
}

func (p *Provider) NewHistogram(name string) *Histogram {
	return &Histogram{
		parent:  p,
		name:    name,
		keyvals: map[string]string{},
	}
}

func (p *Provider) counterAdd(name string, delta float64) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.counters[name] += delta
}

func (p *Provider) gaugeSet(name string, value float64) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.gauges[name] = value
}

func (p *Provider) gaugeAdd(name string, delta float64) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.gauges[name] += delta
}

func (p *Provider) timerObserve(name string, value float64) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.timers[name] = append(p.timers[name], value)
}

func (p *Provider) histogramObserve(name string, value float64) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.timers[name] = append(p.timers[name], value)
}

type Counter struct {
	parent  *Provider
	name    string
	keyvals map[string]string
}

func (c *Counter) With(keyvals ...string) metrics.Counter {
	return &Counter{
		parent:  c.parent,
		name:    c.name,
		keyvals: keyval.Append(c.keyvals, keyvals...),
	}
}

func (c *Counter) Add(delta float64) {
	name := template.Render(c.name, c.keyvals)
	c.parent.counterAdd(name, delta)
}

type Gauge struct {
	parent  *Provider
	name    string
	keyvals map[string]string
}

func (g *Gauge) With(keyvals ...string) metrics.Gauge {
	return &Gauge{
		parent:  g.parent,
		name:    g.name,
		keyvals: keyval.Append(g.keyvals, keyvals...),
	}
}

func (g *Gauge) Add(delta float64) {
	name := template.Render(g.name, g.keyvals)
	g.parent.gaugeAdd(name, delta)
}

func (g *Gauge) Set(value float64) {
	name := template.Render(g.name, g.keyvals)
	g.parent.gaugeSet(name, value)
}

type Timer struct {
	parent  *Provider
	name    string
	keyvals map[string]string
}

func (t *Timer) With(keyvals ...string) metrics.Histogram {
	return &Timer{
		parent:  t.parent,
		name:    t.name,
		keyvals: keyval.Append(t.keyvals, keyvals...),
	}
}

func (t *Timer) Observe(value float64) {
	name := template.Render(t.name, t.keyvals)
	t.parent.timerObserve(name, value)
}

type Histogram struct {
	parent  *Provider
	name    string
	keyvals map[string]string
}

func (h *Histogram) With(keyvals ...string) metrics.Histogram {
	return &Histogram{
		parent:  h.parent,
		name:    h.name,
		keyvals: keyval.Append(h.keyvals, keyvals...),
	}
}

func (h *Histogram) Observe(value float64) {
	name := template.Render(h.name, h.keyvals)
	h.parent.histogramObserve(name, value)
}
