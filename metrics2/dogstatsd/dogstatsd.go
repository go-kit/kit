// Package dogstatsd provides a DogStatsD backend for metrics. It's very similar
// to StatsD, but supports a first-order concept of tags, which we map to Go
// kit's concept of labels. For more details, see the documentation at
// http://docs.datadoghq.com/guides/dogstatsd/.
package dogstatsd

import (
	"fmt"
	"io"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	metrics "github.com/go-kit/kit/metrics2"
	"github.com/go-kit/kit/metrics2/internal/keyval"
)

// Provider constructs and stores DogStatsD metrics. Provider must be
// constructed via NewProvider; the zero value of a provider is not useful.
type Provider struct {
	mtx        sync.RWMutex
	counters   map[nameTags]float64
	gauges     map[nameTags]float64
	histograms map[nameTags][]float64

	// SampleRate, between 0.0 and 1.0 inclusive, instructs the provider to only
	// record and emit a percentage of actual observations. If not set, the
	// default behavior is to record and emit all observations, i.e. a sample
	// rate of 1.0 or 100%.
	SampleRate float64

	// Logger is used to report transport errors.
	// By default, no errors are logged.
	Logger log.Logger
}

// NewProvider returns a new, empty, idle provider. Callers must be sure to
// invoke WriteLoop or SendLoop to actually emit information to a server.
func NewProvider() *Provider {
	return &Provider{
		counters:   map[nameTags]float64{},
		gauges:     map[nameTags]float64{},
		histograms: map[nameTags][]float64{},

		SampleRate: 1.0,
		Logger:     log.NewNopLogger(),
	}
}

// NewCounter returns a Counter whose values are emitted to a DogStatsD backend.
// Only the Name field from the identifier is used.
func (p *Provider) NewCounter(id metrics.Identifier) metrics.Counter {
	return &counter{
		name:    id.Name,
		keyvals: map[string]string{},
		add:     p.counterAdd,
	}
}

// NewGauge returns a Gauge whose values are emitted to a DogStatsD backend.
// Only the Name field from the identifier is used.
func (p *Provider) NewGauge(id metrics.Identifier) metrics.Gauge {
	return &gauge{
		name:    id.Name,
		keyvals: map[string]string{},
		add:     p.gaugeAdd,
		set:     p.gaugeSet,
	}
}

// NewHistogram returns a Histogram whose values are emitted to a DogStatsD
// backend. Only the Name field from the identifier is used.
func (p *Provider) NewHistogram(id metrics.Identifier) metrics.Histogram {
	return &histogram{
		name:    id.Name,
		keyvals: map[string]string{},
		observe: p.histogramObserve,
	}
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
		c, g map[nameTags]float64
		h    map[nameTags][]float64
	)
	{
		p.mtx.Lock()
		c, p.counters = p.counters, map[nameTags]float64{}
		g, p.gauges = p.gauges, map[nameTags]float64{}
		h, p.histograms = p.histograms, map[nameTags][]float64{}
		p.mtx.Unlock()
	}

	// Write the captured data out.
	var (
		sampling = p.sampling()
		count    int64
	)
	for nt, value := range c {
		n, err := fmt.Fprintln(w, nt.name+":"+strconv.FormatFloat(value, 'f', -1, 64)+"|c"+sampling+nt.tags)
		if err != nil {
			return count, err
		}
		count += int64(n)
	}
	for nt, value := range g {
		n, err := fmt.Fprintln(w, nt.name+":"+strconv.FormatFloat(value, 'f', -1, 64)+"|g"+nt.tags)
		if err != nil {
			return count, err
		}
		count += int64(n)
	}
	for nt, values := range h {
		for _, value := range values {
			n, err := fmt.Fprintln(w, nt.name+":"+strconv.FormatFloat(value, 'f', -1, 64)+"|h"+sampling+nt.tags)
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
		return "|@" + strconv.FormatFloat(p.SampleRate, 'f', -1, 64)
	}
	return ""
}

func (p *Provider) counterAdd(nt nameTags, delta float64) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.sampleExec(func() { p.counters[nt] += delta })
}

func (p *Provider) gaugeSet(nt nameTags, value float64) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.sampleExec(func() { p.gauges[nt] = value })
}

func (p *Provider) gaugeAdd(nt nameTags, delta float64) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.sampleExec(func() { p.gauges[nt] += delta })
}

func (p *Provider) histogramObserve(nt nameTags, value float64) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.sampleExec(func() { p.histograms[nt] = append(p.histograms[nt], value) })
}

func (p *Provider) sampleExec(f func()) {
	if p.SampleRate >= 1.0 || p.SampleRate < 0.0 || rand.Float64() < p.SampleRate {
		f()
	}
}

type counter struct {
	name    string
	keyvals map[string]string
	add     func(nt nameTags, delta float64)
}

func (c *counter) With(keyvals ...string) metrics.Counter {
	return &counter{
		name:    c.name,
		keyvals: keyval.Merge(c.keyvals, keyvals...),
		add:     c.add,
	}
}

func (c *counter) Add(delta float64) {
	nt := makeNameTags(c.name, c.keyvals)
	c.add(nt, delta)
}

type gauge struct {
	name    string
	keyvals map[string]string
	add     func(nt nameTags, delta float64)
	set     func(nt nameTags, value float64)
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
	nt := makeNameTags(g.name, g.keyvals)
	g.add(nt, delta)
}

func (g *gauge) Set(value float64) {
	nt := makeNameTags(g.name, g.keyvals)
	g.set(nt, value)
}

type histogram struct {
	name    string
	keyvals map[string]string
	observe func(nt nameTags, value float64)
}

func (h *histogram) With(keyvals ...string) metrics.Histogram {
	return &histogram{
		name:    h.name,
		keyvals: keyval.Merge(h.keyvals, keyvals...),
		observe: h.observe,
	}
}

func (h *histogram) Observe(value float64) {
	nt := makeNameTags(h.name, h.keyvals)
	h.observe(nt, value)
}

type nameTags struct{ name, tags string }

func makeNameTags(name string, keyvals map[string]string) nameTags {
	var tags string
	if len(keyvals) > 0 {
		pairs := make([]string, 0, len(keyvals))
		for k, v := range keyvals {
			pairs = append(pairs, k+":"+v)
		}
		sort.Strings(pairs)
		tags = "|#" + strings.Join(pairs, ",")
	}
	return nameTags{
		name: name,
		tags: tags,
	}
}
