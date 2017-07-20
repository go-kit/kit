// Package statsd provides a StatsD backend for metrics.
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
package statsd

import (
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics2"
	"github.com/go-kit/kit/metrics2/internal/keyval"
	"github.com/go-kit/kit/metrics2/internal/template"
	"github.com/go-kit/kit/util/conn"
)

// Provider constructs and stores StatsD metrics.
// Definitions from https://github.com/b/statsd_spec.
type Provider struct {
	mtx         sync.RWMutex
	counters    map[string]float64
	gauges      map[string]float64
	timers      map[string][]float64
	histograms  map[string][]float64
	floatValues bool
	sampleRate  float64
	logger      log.Logger
}

// ProviderOption changes some behavior of the provider.
// Applied globally to all constructed metrics.
type ProviderOption func(*Provider)

// WithFloatValues instructs the provider to emit values to the StatsD backend
// as floats. Only certain StatsD servers support this mode, so check to make
// sure. If this option isn't provided, all values will be truncated to ints.
func WithFloatValues() ProviderOption {
	return func(p *Provider) { p.floatValues = true }
}

// WithSampleRate instructs the provider to only record, and emit, a percentage
// of actual observations. The primary purpose is to restrict the amount of
// bandwidth used to transmit a report to a StatsD backend.
func WithSampleRate(rate float64) ProviderOption {
	if rate < 0.0 {
		rate = 0.0
	}
	if rate > 1.0 {
		rate = 1.0
	}
	return func(p *Provider) { p.sampleRate = rate }
}

// NewProvider returns a new, empty, idle provider. Callers must be sure to
// invoke WriteLoop or SendLoop to actually emit information to a StatsD
// backend. The logger is used to report transport errors.
func NewProvider(logger log.Logger, options ...ProviderOption) *Provider {
	p := &Provider{
		counters:    map[string]float64{},
		gauges:      map[string]float64{},
		timers:      map[string][]float64{},
		histograms:  map[string][]float64{},
		floatValues: false,
		sampleRate:  1.0,
		logger:      logger,
	}
	for _, option := range options {
		option(p)
	}
	return p
}

// SendLoop connects to a StatsD backend on the given network and address, and
// emits a report every time the passed channel fires. For typical usage, create
// a time.NewTicker and pass the ticker.C channel to this function. The channel
// blocks until the passed channel is closed.
func (p *Provider) SendLoop(c <-chan time.Time, network, address string) {
	p.WriteLoop(c, conn.NewDefaultManager(network, address, p.logger))
}

// WriteLoop writes a report to the passed writer every time the passed channel
// fires. For typical usage, create a time.NewTicker and pass the ticker.C
// channel to this function. The channel blocks until the passed channel is
// closed.
//
// This is a low-level function, primarily intended for testing. Most callers
// should prefer SendLoop.
func (p *Provider) WriteLoop(c <-chan time.Time, w io.Writer) {
	for range c {
		if _, err := p.WriteTo(w); err != nil {
			p.logger.Log("err", err)
		}
	}
}

// WriteTo flushes the buffered contents of the metrics to the passed writer, in
// StatsD format. WriteTo is best-effort and fails fast; observations are lost
// if there's a problem with the write. Clients should be sure to call WriteLoop
// regularly, ideally through the SendLoop or WriteLoop helper methods.
//
// This is a low-level function, primarily intended for testing. Most callers
// should prefer SendLoop.
func (p *Provider) WriteTo(w io.Writer) (int64, error) {
	// Copy the maps and reset them to empty.
	// Do this in a closure to minimize lock time.
	c, g, t, h := func() (c, g map[string]float64, t, h map[string][]float64) {
		p.mtx.Lock()
		defer p.mtx.Unlock()
		c, p.counters = p.counters, map[string]float64{}
		g, p.gauges = p.gauges, map[string]float64{}
		t, p.timers = p.timers, map[string][]float64{}
		h, p.histograms = p.histograms, map[string][]float64{}
		return
	}()
	sampling := p.sampling()
	var count int64
	for name, value := range c {
		n, err := fmt.Fprintln(w, name+":"+p.format(value)+"|c"+sampling)
		if err != nil {
			return count, err
		}
		count += int64(n)
	}
	for name, value := range g {
		n, err := fmt.Fprintln(w, name+":"+p.format(value)+"|g"+sampling)
		if err != nil {
			return count, err
		}
		count += int64(n)
	}
	for name, values := range t {
		for _, value := range values {
			n, err := fmt.Fprintln(w, name+":"+p.format(value)+"|ms"+sampling)
			if err != nil {
				return count, err
			}
			count += int64(n)
		}
	}
	for name, values := range h {
		for _, value := range values {
			n, err := fmt.Fprintln(w, name+":"+p.format(value)+"|h"+sampling)
			if err != nil {
				return count, err
			}
			count += int64(n)
		}
	}
	return count, nil
}

func (p *Provider) sampling() string {
	if p.sampleRate < 1.0 {
		return "|@" + strconv.FormatFloat(p.sampleRate, 'f', 6, 64)
	}
	return ""
}

func (p *Provider) format(value float64) string {
	if p.floatValues {
		return strconv.FormatFloat(value, 'f', 6, 64)
	}
	return strconv.FormatInt(int64(value), 10)
}

// NewCounter returns a Counter whose values are emitted to a StatsD backend.
// The name can include template interpolation to support With; see package
// documentation for details.
func (p *Provider) NewCounter(name string) *Counter {
	return &Counter{
		parent:  p,
		name:    name,
		keyvals: keyval.MakeWith(template.ExtractKeysFrom(name)),
	}
}

// NewGauge returns a Gauge whose values are emitted to a StatsD backend.
// The name can include template interpolation to support With; see package
// documentation for details.
func (p *Provider) NewGauge(name string) *Gauge {
	return &Gauge{
		parent:  p,
		name:    name,
		keyvals: keyval.MakeWith(template.ExtractKeysFrom(name)),
	}
}

// NewTimer returns a Timer whose values are emitted to a StatsD backend. StatsD
// timers map to Go kit histograms, and must take observations in units of
// milliseconds. The name can include template interpolation to support With;
// see package documentation for details.
func (p *Provider) NewTimer(name string) *Timer {
	return &Timer{
		parent:  p,
		name:    name,
		keyvals: keyval.MakeWith(template.ExtractKeysFrom(name)),
	}
}

// NewHistogram returns a Histogram whose values are emitted to a StatsD
// backend. The name can include template interpolation to support With; see
// package documentation for details.
func (p *Provider) NewHistogram(name string) *Histogram {
	return &Histogram{
		parent:  p,
		name:    name,
		keyvals: keyval.MakeWith(template.ExtractKeysFrom(name)),
	}
}

func (p *Provider) counterAdd(name string, delta float64) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.sampleExec(func() { p.counters[name] += delta })
}

func (p *Provider) gaugeSet(name string, value float64) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.sampleExec(func() { p.gauges[name] = value })
}

func (p *Provider) gaugeAdd(name string, delta float64) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.sampleExec(func() { p.gauges[name] += delta })
}

func (p *Provider) timerObserve(name string, value float64) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.sampleExec(func() { p.timers[name] = append(p.timers[name], value) })
}

func (p *Provider) histogramObserve(name string, value float64) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.sampleExec(func() { p.histograms[name] = append(p.histograms[name], value) })
}

func (p *Provider) sampleExec(f func()) {
	if p.sampleRate >= 1.0 || rand.Float64() < p.sampleRate {
		f()
	}
}

// Counter is a StatsD counter object. Counters must be constructed via the
// Provider; the zero value of a Counter is not useful.
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
	c.parent.counterAdd(name, delta)
}

// Gauge is a StatsD gauge object. Gauges must be constructed via the Provider;
// the zero value of a Gauge is not useful.
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

// Add implements Gauge.
func (g *Gauge) Add(delta float64) {
	name := template.Render(g.name, g.keyvals)
	g.parent.gaugeAdd(name, delta)
}

// Set implements Gauge.
func (g *Gauge) Set(value float64) {
	name := template.Render(g.name, g.keyvals)
	g.parent.gaugeSet(name, value)
}

// Timer is a StatsD timer object, modeled as a Go kit histogram. Timer
// observations are expected to be in units of milliseconds. Timers must be
// constructed via the Provider; the zero value of a Timer is not useful.
type Timer struct {
	parent  *Provider
	name    string
	keyvals map[string]string
}

// With implements Histogram.
func (t *Timer) With(keyvals ...string) metrics.Histogram {
	return &Timer{
		parent:  t.parent,
		name:    t.name,
		keyvals: keyval.Merge(t.keyvals, keyvals...),
	}
}

// Observe implements Histogram.
func (t *Timer) Observe(value float64) {
	name := template.Render(t.name, t.keyvals)
	t.parent.timerObserve(name, value)
}

// Histogram is a StatsD histogram object. Histogram observations are unitless
// in the protocol. Histograms must be constructed via the Provider; the zero
// value of a Histogram is not useful.
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
	h.parent.histogramObserve(name, value)
}
