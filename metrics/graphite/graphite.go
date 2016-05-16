// Package graphite implements a Graphite backend for package metrics. Metrics
// will be emitted to a Graphite server in the plaintext protocol which looks
// like:
//
//   "<metric path> <metric value> <metric timestamp>"
//
// See http://graphite.readthedocs.io/en/latest/feeding-carbon.html#the-plaintext-protocol.
// The current implementation ignores fields.
package graphite

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"net"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/codahale/hdrhistogram"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
)

// Emitter will keep track of all metrics and, once started, will emit the
// metrics via the Flush method to the given address.
type Emitter struct {
	mtx        sync.Mutex
	prefix     string
	mgr        *manager
	counters   []*counter
	histograms []*windowedHistogram
	gauges     []*gauge
	logger     log.Logger
	quitc      chan chan struct{}
}

// NewEmitter will return an Emitter that will prefix all metrics names with the
// given prefix. Once started, it will attempt to create a connection with the
// given network and address via `net.Dial` and periodically post metrics to the
// connection in the Graphite plaintext protocol.
func NewEmitter(network, address string, metricsPrefix string, flushInterval time.Duration, logger log.Logger) *Emitter {
	return NewEmitterDial(net.Dial, network, address, metricsPrefix, flushInterval, logger)
}

// NewEmitterDial is the same as NewEmitter, but allows you to specify your own
// Dialer function. This is primarily useful for tests.
func NewEmitterDial(dialer Dialer, network, address string, metricsPrefix string, flushInterval time.Duration, logger log.Logger) *Emitter {
	e := &Emitter{
		prefix: metricsPrefix,
		mgr:    newManager(dialer, network, address, time.After, logger),
		logger: logger,
		quitc:  make(chan chan struct{}),
	}
	go e.loop(flushInterval)
	return e
}

// NewCounter returns a Counter whose value will be periodically emitted in
// a Graphite-compatible format once the Emitter is started. Fields are ignored.
func (e *Emitter) NewCounter(name string) metrics.Counter {
	e.mtx.Lock()
	defer e.mtx.Unlock()
	c := &counter{name, 0}
	e.counters = append(e.counters, c)
	return c
}

// NewHistogram is taken from http://github.com/codahale/metrics. It returns a
// windowed HDR histogram which drops data older than five minutes.
//
// The histogram exposes metrics for each passed quantile as gauges. Quantiles
// should be integers in the range 1..99. The gauge names are assigned by using
// the passed name as a prefix and appending "_pNN" e.g. "_p50".
//
// The values of this histogram will be periodically emitted in a
// Graphite-compatible format once the Emitter is started. Fields are ignored.
func (e *Emitter) NewHistogram(name string, minValue, maxValue int64, sigfigs int, quantiles ...int) (metrics.Histogram, error) {
	gauges := map[int]metrics.Gauge{}
	for _, quantile := range quantiles {
		if quantile <= 0 || quantile >= 100 {
			return nil, fmt.Errorf("invalid quantile %d", quantile)
		}
		gauges[quantile] = e.gauge(fmt.Sprintf("%s_p%02d", name, quantile))
	}
	h := newWindowedHistogram(name, minValue, maxValue, sigfigs, gauges, e.logger)

	e.mtx.Lock()
	defer e.mtx.Unlock()
	e.histograms = append(e.histograms, h)
	return h, nil
}

// NewTimeHistogram returns a TimeHistogram wrapper around the windowed HDR
// histrogram provided by this package.
func (e *Emitter) NewTimeHistogram(name string, unit time.Duration, minValue, maxValue int64, sigfigs int, quantiles ...int) (metrics.TimeHistogram, error) {
	h, err := e.NewHistogram(name, minValue, maxValue, sigfigs, quantiles...)
	if err != nil {
		return nil, err
	}
	return metrics.NewTimeHistogram(unit, h), nil
}

// NewGauge returns a Gauge whose value will be periodically emitted in a
// Graphite-compatible format once the Emitter is started. Fields are ignored.
func (e *Emitter) NewGauge(name string) metrics.Gauge {
	e.mtx.Lock()
	defer e.mtx.Unlock()
	return e.gauge(name)
}

func (e *Emitter) gauge(name string) metrics.Gauge {
	g := &gauge{name, 0}
	e.gauges = append(e.gauges, g)
	return g
}

func (e *Emitter) loop(d time.Duration) {
	ticker := time.NewTicker(d)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			e.Flush()

		case q := <-e.quitc:
			e.Flush()
			close(q)
			return
		}
	}
}

// Stop will flush the current metrics and close the active connection. Calling
// stop more than once is a programmer error.
func (e *Emitter) Stop() {
	q := make(chan struct{})
	e.quitc <- q
	<-q
}

// Flush will write the current metrics to the Emitter's connection in the
// Graphite plaintext protocol.
func (e *Emitter) Flush() {
	e.mtx.Lock() // one flush at a time
	defer e.mtx.Unlock()

	conn := e.mgr.take()
	if conn == nil {
		e.logger.Log("during", "flush", "err", "connection unavailable")
		return
	}

	err := e.flush(conn)
	if err != nil {
		e.logger.Log("during", "flush", "err", err)
	}
	e.mgr.put(err)
}

func (e *Emitter) flush(conn io.Writer) error {
	w := bufio.NewWriter(conn)

	for _, c := range e.counters {
		fmt.Fprintf(w, "%s.%s.count %d %d\n", e.prefix, c.Name(), c.count, time.Now().Unix())
	}

	for _, h := range e.histograms {
		hist := h.hist.Merge()
		now := time.Now().Unix()
		fmt.Fprintf(w, "%s.%s.count %d %d\n", e.prefix, h.Name(), hist.TotalCount(), now)
		fmt.Fprintf(w, "%s.%s.min %d %d\n", e.prefix, h.Name(), hist.Min(), now)
		fmt.Fprintf(w, "%s.%s.max %d %d\n", e.prefix, h.Name(), hist.Max(), now)
		fmt.Fprintf(w, "%s.%s.mean %.2f %d\n", e.prefix, h.Name(), hist.Mean(), now)
		fmt.Fprintf(w, "%s.%s.std-dev %.2f %d\n", e.prefix, h.Name(), hist.StdDev(), now)
	}

	for _, g := range e.gauges {
		fmt.Fprintf(w, "%s.%s %.2f %d\n", e.prefix, g.Name(), g.Get(), time.Now().Unix())
	}

	return w.Flush()
}

type counter struct {
	key   string
	count uint64
}

func (c *counter) Name() string { return c.key }

func (c *counter) With(metrics.Field) metrics.Counter { return c }

func (c *counter) Add(delta uint64) { atomic.AddUint64(&c.count, delta) }

type gauge struct {
	key   string
	value uint64 // math.Float64bits
}

func (g *gauge) Name() string { return g.key }

func (g *gauge) With(metrics.Field) metrics.Gauge { return g }

func (g *gauge) Add(delta float64) {
	for {
		old := atomic.LoadUint64(&g.value)
		new := math.Float64bits(math.Float64frombits(old) + delta)
		if atomic.CompareAndSwapUint64(&g.value, old, new) {
			return
		}
	}
}

func (g *gauge) Set(value float64) {
	atomic.StoreUint64(&g.value, math.Float64bits(value))
}

func (g *gauge) Get() float64 {
	return math.Float64frombits(atomic.LoadUint64(&g.value))
}

type windowedHistogram struct {
	mtx  sync.Mutex
	hist *hdrhistogram.WindowedHistogram

	name   string
	gauges map[int]metrics.Gauge
	logger log.Logger
}

// newWindowedHistogram is taken from http://github.com/codahale/metrics. It
// returns a windowed HDR histogram which drops data older than five minutes.
//
// The histogram exposes metrics for each passed quantile as gauges. Users are
// expected to provide their own set of Gauges for quantiles to make this
// Histogram work across multiple metrics providers.
func newWindowedHistogram(name string, minValue, maxValue int64, sigfigs int, quantiles map[int]metrics.Gauge, logger log.Logger) *windowedHistogram {
	h := &windowedHistogram{
		hist:   hdrhistogram.NewWindowed(5, minValue, maxValue, sigfigs),
		name:   name,
		gauges: quantiles,
		logger: logger,
	}
	go h.rotateLoop(1 * time.Minute)
	return h
}

func (h *windowedHistogram) Name() string { return h.name }

func (h *windowedHistogram) With(metrics.Field) metrics.Histogram { return h }

func (h *windowedHistogram) Observe(value int64) {
	h.mtx.Lock()
	err := h.hist.Current.RecordValue(value)
	h.mtx.Unlock()

	if err != nil {
		h.logger.Log("err", err, "msg", "unable to record histogram value")
		return
	}

	for q, gauge := range h.gauges {
		gauge.Set(float64(h.hist.Current.ValueAtQuantile(float64(q))))
	}
}

func (h *windowedHistogram) Distribution() ([]metrics.Bucket, []metrics.Quantile) {
	bars := h.hist.Merge().Distribution()
	buckets := make([]metrics.Bucket, len(bars))
	for i, bar := range bars {
		buckets[i] = metrics.Bucket{
			From:  bar.From,
			To:    bar.To,
			Count: bar.Count,
		}
	}
	quantiles := make([]metrics.Quantile, 0, len(h.gauges))
	for quantile, gauge := range h.gauges {
		quantiles = append(quantiles, metrics.Quantile{
			Quantile: quantile,
			Value:    int64(gauge.Get()),
		})
	}
	sort.Sort(quantileSlice(quantiles))
	return buckets, quantiles
}

func (h *windowedHistogram) rotateLoop(d time.Duration) {
	for range time.Tick(d) {
		h.mtx.Lock()
		h.hist.Rotate()
		h.mtx.Unlock()
	}
}

type quantileSlice []metrics.Quantile

func (a quantileSlice) Len() int           { return len(a) }
func (a quantileSlice) Less(i, j int) bool { return a[i].Quantile < a[j].Quantile }
func (a quantileSlice) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
