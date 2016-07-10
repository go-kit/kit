// Package dogstatsd provides a DogStatsD backend for package metrics. It's very
// similar to statsd, but supports arbitrary tags per-metric, which map to Go
// kit's label values. So, while label values are no-ops in statsd, they are
// supported here. For more details, see the documentation at
// http://docs.datadoghq.com/guides/dogstatsd/.
//
// This package batches observations and emits them on some schedule to the
// remote server. This is useful even if you connect to your DogStatsD server
// over UDP. Emitting one network packet per observation can quickly overwhelm
// even the fastest internal network. Batching allows you to more linearly scale
// with growth.
//
// Typically you'll create a Dogstatsd object in your main function.
//
//    d, stop := New("myprefix.", "udp", "dogstatsd:8125", time.Second, log.NewNopLogger())
//    defer stop()
//
// Then, create the metrics that your application will track from that object.
// Pass them as dependencies to the component that needs them; don't place them
// in the global scope.
//
//    requests := d.NewCounter("requests")
//    foo := NewFoo(store, logger, requests)
//
// Invoke them in your components when you have something to instrument.
//
//    f.requests.Add(1)
//
package dogstatsd

import (
	"fmt"
	"io"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics2"
	"github.com/go-kit/kit/metrics2/generic"
	"github.com/go-kit/kit/metrics2/statsd"
	"github.com/go-kit/kit/util/conn"
)

// Dogstatsd is a store for metrics that will be reported to a DogStatsD server.
// Create a Dogstatsd object, use it to create metrics objects, and pass those
// objects as dependencies to the components that will use them.
type Dogstatsd struct {
	mtx        sync.RWMutex
	prefix     string
	counters   map[string]*generic.Counter
	gauges     map[string]*generic.Gauge
	histograms map[string]*Histogram
	timings    map[string]*statsd.Timing
	sets       map[string]*Set
	logger     log.Logger
}

// NewRaw creates a Dogstatsd object. By default the metrics will not be emitted
// anywhere. Use WriteTo to flush the metrics once, or FlushTo (in a separate
// goroutine) to flush them on a regular schedule, or use the New constructor to
// set up the object and flushing at the same time.
func NewRaw(prefix string, logger log.Logger) *Dogstatsd {
	return &Dogstatsd{
		prefix:     prefix,
		counters:   map[string]*generic.Counter{},
		gauges:     map[string]*generic.Gauge{},
		histograms: map[string]*Histogram{},
		timings:    map[string]*statsd.Timing{},
		sets:       map[string]*Set{},
		logger:     logger,
	}
}

// New creates a Statsd object that flushes all metrics in the DogStatsD format
// every flushInterval to the network and address. Use the returned stop
// function to terminate the flushing goroutine.
func New(prefix string, network, address string, flushInterval time.Duration, logger log.Logger) (res *Dogstatsd, stop func()) {
	d := NewRaw(prefix, logger)
	manager := conn.NewDefaultManager(network, address, logger)
	ticker := time.NewTicker(flushInterval)
	go d.FlushTo(manager, ticker)
	return d, ticker.Stop
}

// NewCounter returns a counter metric with the given name. Adds are buffered
// until the underlying Statsd object is flushed.
func (d *Dogstatsd) NewCounter(name string) *generic.Counter {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	c := generic.NewCounter()
	d.counters[d.prefix+name] = c
	return c
}

// NewGauge returns a gauge metric with the given name.
func (d *Dogstatsd) NewGauge(name string) *generic.Gauge {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	g := generic.NewGauge()
	d.gauges[d.prefix+name] = g
	return g
}

// NewHistogram returns a histogram metric with the given name and sample rate.
func (d *Dogstatsd) NewHistogram(name string, sampleRate float64) *Histogram {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	h := newHistogram(sampleRate)
	d.histograms[d.prefix+name] = h
	return h
}

// NewTiming returns a StatsD timing metric (DogStatsD documentation calls them
// Timers) with the given name, unit (e.g. "ms") and sample rate. Pass a sample
// rate of 1.0 or greater to disable sampling. Sampling is done at observation
// time. Observations are buffered until the underlying statsd object is
// flushed.
func (d *Dogstatsd) NewTiming(name, unit string, sampleRate float64) *statsd.Timing {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	t := statsd.NewTiming(unit, sampleRate)
	d.timings[d.prefix+name] = t
	return t
}

// NewSet returns a DogStatsD set with the given name.
func (d *Dogstatsd) NewSet(name string) *Set {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	s := newSet()
	d.sets[d.prefix+name] = s
	return s
}

// FlushTo invokes WriteTo to the writer every time the ticker fires. FlushTo
// blocks until the ticker is stopped. Most users won't need to call this method
// directly, and should prefer to use the New constructor.
func (d *Dogstatsd) FlushTo(w io.Writer, ticker *time.Ticker) {
	for range ticker.C {
		if _, err := d.WriteTo(w); err != nil {
			d.logger.Log("during", "flush", "err", err)
		}
	}
}

// WriteTo dumps the current state of all of the metrics to the given writer in
// the DogStatsD format. Each metric has its current value(s) written in
// sequential calls to Write. Counters and gauges are dumped with their current
// values; counters are reset. Histograms and timers have all of their
// (potentially sampled) observations dumped, and are reset. Sets have all of
// their observations dumped and are reset. Clients probably shouldn't invoke
// this method directly, and should prefer using FlushTo, or the New
// constructor.
func (d *Dogstatsd) WriteTo(w io.Writer) (int64, error) {
	d.mtx.RLock()
	defer d.mtx.RUnlock()
	var (
		n     int
		err   error
		count int64
	)
	for name, c := range d.counters {
		value := c.ValueReset()
		tv := tagValues(c.LabelValues())
		n, err = fmt.Fprintf(w, "%s:%f|c%s\n", name, value, tv)
		count += int64(n)
		if err != nil {
			return count, err
		}
	}
	for name, g := range d.gauges {
		value := g.Value()
		tv := tagValues(g.LabelValues())
		n, err := fmt.Fprintf(w, "%s:%f|g%s\n", name, value, tv)
		count += int64(n)
		if err != nil {
			return count, err
		}
	}
	for name, h := range d.histograms {
		sv := sampling(h.sampleRate)
		tv := tagValues(h.lvs)
		for _, value := range h.values {
			n, err = fmt.Fprintf(w, "%s:%f|h%s%s\n", name, value, sv, tv)
			count += int64(n)
			if err != nil {
				return count, err
			}
		}
	}
	for name, t := range d.timings {
		un := t.Unit()
		sv := sampling(t.SampleRate())
		tv := tagValues(t.LabelValues())
		for _, value := range t.Values() {
			n, err = fmt.Fprintf(w, "%s:%d|%s%s%s\n", name, value, un, sv, tv)
			count += int64(n)
			if err != nil {
				return count, err
			}
		}
	}
	for name, s := range d.sets {
		for _, value := range s.Values() {
			n, err = fmt.Fprintf(w, "%s:%s|s\n", name, value)
			count += int64(n)
			if err != nil {
				return count, err
			}
		}
	}
	return count, nil
}

// Histogram is a denormalized collection of observed values. A general form of
// a StatsD Timing. Create histograms through the Dogstatsd object.
type Histogram struct {
	mtx        sync.Mutex
	sampleRate float64
	values     []float64
	lvs        []string
}

func newHistogram(sampleRate float64) *Histogram {
	return &Histogram{
		sampleRate: sampleRate,
	}
}

// With implements metrics.Histogram.
func (h *Histogram) With(labelValues ...string) metrics.Histogram {
	if len(labelValues)%2 != 0 {
		labelValues = append(labelValues, generic.LabelValueUnknown)
	}
	return &Histogram{
		sampleRate: h.sampleRate,
		values:     h.values,
		lvs:        append(h.lvs, labelValues...),
	}
}

// Observe implements metrics.Histogram. Values are simply aggregated into memory.
// If the sample rate is less than 1.0, observations may be dropped.
func (h *Histogram) Observe(value float64) {
	if h.sampleRate < 1.0 && rand.Float64() > h.sampleRate {
		return
	}
	h.mtx.Lock()
	defer h.mtx.Unlock()
	h.values = append(h.values, value)
}

// Values returns the observed values since the last call to Values. This method
// clears the internal state of the Histogram; better get those values somewhere
// safe!
func (h *Histogram) Values() []float64 {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	res := h.values
	h.values = []float64{}
	return res
}

// Set is a DogStatsD-specific metric for tracking unique identifiers.
// Create sets through the Dogstatsd object.
type Set struct {
	mtx    sync.Mutex
	values map[string]struct{}
}

func newSet() *Set {
	return &Set{
		values: map[string]struct{}{},
	}
}

// Observe adds the value to the set.
func (s *Set) Observe(value string) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.values[value] = struct{}{}
}

// Values returns the unique observed values since the last call to Values. This
// method clears the internal state of the Set; better get those values
// somewhere safe!
func (s *Set) Values() []string {
	res := make([]string, 0, len(s.values))
	for value := range s.values {
		res = append(res, value)
	}
	s.values = map[string]struct{}{} // TODO(pb): if GC is a problem, this can be improved
	return res
}

func sampling(r float64) string {
	var sv string
	if r < 1.0 {
		sv = fmt.Sprintf("|@%f", r)
	}
	return sv
}

func tagValues(labelValues []string) string {
	if len(labelValues) == 0 {
		return ""
	}
	if len(labelValues)%2 != 0 {
		panic("tagValues received a labelValues with an odd number of strings")
	}
	pairs := make([]string, 0, len(labelValues)/2)
	for i := 0; i < len(labelValues); i += 2 {
		pairs = append(pairs, labelValues[i]+":"+labelValues[i+1])
	}
	return "|#" + strings.Join(pairs, ",")
}
