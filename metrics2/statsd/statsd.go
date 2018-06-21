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

// Provider constructs and stores StatsD metrics. Definitions from
// https://github.com/b/statsd_spec. Provider must be constructed via
// NewProvider; the zero value of a provider is not useful.
type Provider struct {
	mtx        sync.RWMutex
	counters   map[string]float64
	gauges     map[string]float64
	timers     map[string][]float64
	histograms map[string][]float64

	// EmitFloatValues instructs the provider to emit values to the StatsD
	// server as floats. Only certain StatsD servers support this mode, so check
	// to make sure. By default, the provider will truncate all values to ints.
	EmitFloatValues bool

	// SampleRate, between 0.0 and 1.0 inclusive, instructs the provider to only
	// record and emit a percentage of actual observations. The primary purpose
	// is to restrict the amount of bandwidth used to transmit reports to a
	// server. If not set, the default behavior is to record and emit all
	// observations, i.e. a sample rate of 1.0 or 100%.
	SampleRate float64

	// Logger is used to report transport errors.
	// By default, no errors are logged.
	Logger log.Logger
}

// NewProvider returns a new, empty, idle provider. Callers must be sure to
// invoke WriteLoop or SendLoop to actually emit information to a server.
func NewProvider() *Provider {
	return &Provider{
		counters:   map[string]float64{},
		gauges:     map[string]float64{},
		timers:     map[string][]float64{},
		histograms: map[string][]float64{},

		EmitFloatValues: false,
		SampleRate:      1.0,
		Logger:          log.NewNopLogger(),
	}
}

// SendLoop connects to a StatsD backend on the given network and address, and
// emits a report every time the passed channel fires. For typical usage, create
// a time.NewTicker and pass the ticker.C channel to this function. The channel
// blocks until the passed channel is closed.
func (p *Provider) SendLoop(c <-chan time.Time, network, address string) {
	p.WriteLoop(c, conn.NewDefaultManager(network, address, p.Logger))
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
			p.Logger.Log("err", err)
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
	var (
		c, g map[string]float64
		t, h map[string][]float64
	)
	{
		p.mtx.Lock()
		c, p.counters = p.counters, map[string]float64{}
		g, p.gauges = p.gauges, map[string]float64{}
		t, p.timers = p.timers, map[string][]float64{}
		h, p.histograms = p.histograms, map[string][]float64{}
		p.mtx.Unlock()
	}

	// Write the captured data out.
	var (
		sampling = p.sampling()
		count    int64
	)
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
	if 0.0 < p.SampleRate && p.SampleRate < 1.0 {
		return "|@" + strconv.FormatFloat(p.SampleRate, 'f', 6, 64)
	}
	return ""
}

func (p *Provider) format(value float64) string {
	if p.EmitFloatValues {
		return strconv.FormatFloat(value, 'f', 6, 64)
	}
	return strconv.FormatInt(int64(value), 10)
}

// NewCounter returns a Counter whose values are emitted to a StatsD backend.
//
// Only the NameTemplate field from the identifier is used. It can include
// template interpolation to support With; see package documentation for
// details.
func (p *Provider) NewCounter(id metrics.Identifier) metrics.Counter {
	return &counter{
		name:    id.NameTemplate,
		keyvals: keyval.MakeWith(template.ExtractKeysFrom(id.NameTemplate)),
		add:     p.counterAdd,
	}
}

// NewGauge returns a Gauge whose values are emitted to a StatsD backend.
//
// Only the NameTemplate field from the identifier is used. It can include
// template interpolation to support With; see package documentation for
// details.
func (p *Provider) NewGauge(id metrics.Identifier) metrics.Gauge {
	return &gauge{
		name:    id.NameTemplate,
		keyvals: keyval.MakeWith(template.ExtractKeysFrom(id.NameTemplate)),
		add:     p.gaugeAdd,
		set:     p.gaugeSet,
	}
}

// NewTimer returns a Timer whose values are emitted to a StatsD backend. StatsD
// timers map to Go kit histograms, and must take observations in units of
// milliseconds.
//
// Only the NameTemplate field from the identifier is used. It can include
// template interpolation to support With; see package documentation for
// details.
func (p *Provider) NewTimer(id metrics.Identifier) metrics.Histogram {
	return &histogram{
		name:    id.NameTemplate,
		keyvals: keyval.MakeWith(template.ExtractKeysFrom(id.NameTemplate)),
		observe: p.timerObserve,
	}
}

// NewHistogram returns a Histogram whose values are emitted to a StatsD
// backend.
//
// Only the NameTemplate field from the identifier is used. It can include
// template interpolation to support With; see package documentation for
// details.
func (p *Provider) NewHistogram(id metrics.Identifier) metrics.Histogram {
	return &histogram{
		name:    id.NameTemplate,
		keyvals: keyval.MakeWith(template.ExtractKeysFrom(id.NameTemplate)),
		observe: p.histogramObserve,
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
	if p.SampleRate >= 1.0 || p.SampleRate < 0.0 || rand.Float64() < p.SampleRate {
		f()
	}
}

type counter struct {
	name    string
	keyvals map[string]string
	add     func(name string, delta float64)
}

func (c *counter) With(keyvals ...string) metrics.Counter {
	return &counter{
		name:    c.name,
		keyvals: keyval.Merge(c.keyvals, keyvals...),
		add:     c.add,
	}
}

func (c *counter) Add(delta float64) {
	name := template.Render(c.name, c.keyvals)
	c.add(name, delta)
}

type gauge struct {
	name    string
	keyvals map[string]string
	add     func(name string, delta float64)
	set     func(name string, value float64)
}

func (g *gauge) With(keyvals ...string) metrics.Gauge {
	return &gauge{
		name:    g.name,
		keyvals: keyval.Merge(g.keyvals, keyvals...),
		add:     g.add,
		set:     g.set,
	}
}

func (g *gauge) Add(delta float64) {
	name := template.Render(g.name, g.keyvals)
	g.add(name, delta)
}

func (g *gauge) Set(value float64) {
	name := template.Render(g.name, g.keyvals)
	g.set(name, value)
}

type histogram struct {
	name    string
	keyvals map[string]string
	observe func(name string, value float64)
}

func (h *histogram) With(keyvals ...string) metrics.Histogram {
	return &histogram{
		name:    h.name,
		keyvals: keyval.Merge(h.keyvals, keyvals...),
		observe: h.observe,
	}
}

func (h *histogram) Observe(value float64) {
	name := template.Render(h.name, h.keyvals)
	h.observe(name, value)
}
