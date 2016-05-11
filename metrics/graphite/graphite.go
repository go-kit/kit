// Package graphite implements a graphite backend for package metrics.
//
// The current implementation ignores fields.
package graphite

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"sort"
	"sync"
	"time"

	"sync/atomic"

	"github.com/codahale/hdrhistogram"
	"github.com/go-kit/kit/metrics"
)

// Emitter will keep track of all metrics and, once started,
// will emit the metrics via the Flush method to the given io.Writer.
type Emitter interface {
	NewCounter(string) metrics.Counter
	NewHistogram(string, int64, int64, int, ...int) metrics.Histogram
	NewTimeHistogram(string, time.Duration, int64, int64, int, ...int) metrics.TimeHistogram
	NewGauge(string) metrics.Gauge

	Start(time.Duration)
	Flush() error
}

type emitter struct {
	addr   *net.TCPAddr
	prefix string

	metricMu   *sync.Mutex
	counters   []*counter
	histograms []*windowedHistogram
	gauges     []*gauge
}

// NewEmitter will return an Emitter that will prefix all
// metrics names with the given prefix. Once started, it will attempt to create
// a TCP connection with the given address and most metrics to the connection
// in a Graphite-compatible format.
func NewEmitter(addr *net.TCPAddr, prefix string) Emitter {
	e := &emitter{
		addr, prefix, &sync.Mutex{},
		[]*counter{}, []*windowedHistogram{}, []*gauge{},
	}

	return e
}

// NewCounter returns a Counter whose value will be periodically emitted in
// a Graphite-compatible format once the Emitter is started. Fields are ignored.
func (e *emitter) NewCounter(name string) metrics.Counter {
	// only one flush at a time
	e.metricMu.Lock()
	defer e.metricMu.Unlock()
	c := &counter{name, 0}
	e.counters = append(e.counters, c)
	return c
}

// NewHistogram is taken from http://github.com/codahale/metrics. It returns a
// windowed HDR histogram which drops data older than five minutes.
//
// The histogram exposes metrics for each passed quantile as gauges. Quantiles
// should be integers in the range 1..99. The gauge names are assigned by
// using the passed name as a prefix and appending "_pNN" e.g. "_p50".
//
// The values of this histogram will be periodically emitted in a Graphite-compatible
// format once the Emitter is started. Fields are ignored.
func (e *emitter) NewHistogram(name string, minValue, maxValue int64, sigfigs int, quantiles ...int) metrics.Histogram {
	// only one flush at a time
	e.metricMu.Lock()
	defer e.metricMu.Unlock()

	gauges := map[int]metrics.Gauge{}
	for _, quantile := range quantiles {
		if quantile <= 0 || quantile >= 100 {
			panic(fmt.Sprintf("invalid quantile %d", quantile))
		}
		gauges[quantile] = e.gauge(fmt.Sprintf("%s_p%02d", name, quantile))
	}
	h := newWindowedHistogram(name, minValue, maxValue, sigfigs, gauges)
	e.histograms = append(e.histograms, h)
	return h
}

// NewTimeHistogram returns a TimeHistogram wrapper around the windowed
// HDR histrogram provided by this package.
func (e *emitter) NewTimeHistogram(name string, unit time.Duration, minValue, maxValue int64, sigfigs int, quantiles ...int) metrics.TimeHistogram {
	h := e.NewHistogram(name, minValue, maxValue, sigfigs, quantiles...)
	return metrics.NewTimeHistogram(unit, h)
}

// NewGauge returns a Gauge whose value will be periodically emitted in
// a Graphite-compatible format once the Emitter is started. Fields are ignored.
func (e *emitter) NewGauge(name string) metrics.Gauge {
	// only one flush at a time
	e.metricMu.Lock()
	defer e.metricMu.Unlock()
	return e.gauge(name)
}

func (e *emitter) gauge(name string) metrics.Gauge {
	g := &gauge{name, 0}
	e.gauges = append(e.gauges, g)
	return g
}

// Start will kick off a background goroutine to
// call Flush once every interval.
func (e *emitter) Start(interval time.Duration) {
	go func() {
		t := time.Tick(interval)
		for range t {
			err := e.Flush()
			if err != nil {
				log.Print("error: could not dial graphite host: ", err)
				continue
			}
		}
	}()
}

// Flush will attempt to create a connection with the given address
// and write the current metrics to it in a Graphite-compatible format.
//
// Users can call this method on process shutdown to ensure
// the current metrics are pushed to Graphite.
func (e *emitter) Flush() error {
	// open connection
	conn, err := net.DialTCP("tcp", nil, e.addr)
	if err != nil {
		return err
	}

	// flush stats to connection
	e.flush(conn)

	// close connection
	conn.Close()
	return nil
}

func (e *emitter) flush(conn io.Writer) {
	// only one flush at a time
	e.metricMu.Lock()
	defer e.metricMu.Unlock()

	// buffer the writer and make sure to flush it
	w := bufio.NewWriter(conn)
	defer w.Flush()

	now := time.Now().Unix()

	// emit counter stats
	for _, c := range e.counters {
		fmt.Fprintf(w, "%s.%s.count %d %d\n", e.prefix, c.Name(), c.count, now)
	}

	// emit histogram specific stats
	for _, h := range e.histograms {
		hist := h.hist.Merge()
		fmt.Fprintf(w, "%s.%s.count %d %d\n", e.prefix, h.Name(), hist.TotalCount(), now)
		fmt.Fprintf(w, "%s.%s.min %d %d\n", e.prefix, h.Name(), hist.Min(), now)
		fmt.Fprintf(w, "%s.%s.max %d %d\n", e.prefix, h.Name(), hist.Max(), now)
		fmt.Fprintf(w, "%s.%s.mean %.2f %d\n", e.prefix, h.Name(), hist.Mean(), now)
		fmt.Fprintf(w, "%s.%s.std-dev %.2f %d\n", e.prefix, h.Name(), hist.StdDev(), now)
	}

	// emit gauge stats (which can include some histogram quantiles)
	for _, g := range e.gauges {
		fmt.Fprintf(w, "%s.%s %.2f %d\n", e.prefix, g.Name(), g.Get(), now)
	}
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
	mu   sync.Mutex
	hist *hdrhistogram.WindowedHistogram

	name   string
	gauges map[int]metrics.Gauge
}

// NewWindowedHistogram is taken from http://github.com/codahale/metrics. It returns a
// windowed HDR histogram which drops data older than five minutes.
//
// The histogram exposes metrics for each passed quantile as gauges. Users are expected
// to provide their own set of Gauges for quantiles to make this Histogram work across multiple
// metrics providers.
func newWindowedHistogram(name string, minValue, maxValue int64, sigfigs int, quantiles map[int]metrics.Gauge) *windowedHistogram {
	h := &windowedHistogram{
		hist:   hdrhistogram.NewWindowed(5, minValue, maxValue, sigfigs),
		name:   name,
		gauges: quantiles,
	}
	go h.rotateLoop(1 * time.Minute)
	return h
}

func (h *windowedHistogram) Name() string                         { return h.name }
func (h *windowedHistogram) With(metrics.Field) metrics.Histogram { return h }

func (h *windowedHistogram) Observe(value int64) {
	h.mu.Lock()
	err := h.hist.Current.RecordValue(value)
	h.mu.Unlock()

	if err != nil {
		panic(err.Error())
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
		h.mu.Lock()
		h.hist.Rotate()
		h.mu.Unlock()
	}
}

type quantileSlice []metrics.Quantile

func (a quantileSlice) Len() int           { return len(a) }
func (a quantileSlice) Less(i, j int) bool { return a[i].Quantile < a[j].Quantile }
func (a quantileSlice) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
