// Package graphite implements a Graphite backend for package metrics. Metrics
// will be emitted to a Graphite server in the plaintext protocol
// (http://graphite.readthedocs.io/en/latest/feeding-carbon.html#the-plaintext-protocol)
// which looks like:
//   "<metric path> <metric value> <metric timestamp>"
//
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

// Emitter will keep track of all metrics and, once started,
// will emit the metrics via the Flush method to the given address.
type Emitter interface {
	NewCounter(name string) metrics.Counter
	NewHistogram(name string, min int64, max int64, sigfigs int, quantiles ...int) (metrics.Histogram, error)
	NewTimeHistogram(name string, unit time.Duration, min int64, max int64, sigfigs int, quantiles ...int) (metrics.TimeHistogram, error)
	NewGauge(name string) metrics.Gauge

	Start(reportInvterval time.Duration) error
	Flush()
	Stop() error
}

type emitter struct {
	prefix string

	addr  string
	tcp   bool
	conn  net.Conn
	start sync.Once
	stop  chan bool

	mtx        sync.Mutex
	counters   []*counter
	histograms []*windowedHistogram
	gauges     []*gauge

	logger log.Logger
}

// NewEmitter will return an Emitter that will prefix all
// metrics names with the given prefix. Once started, it will attempt to create
// a TCP or a UDP connection with the given address and periodically post
// metrics to the connection in the Graphite plaintext protocol.
// If the provided `tcp` parameter is false, a UDP connection will be used.
func NewEmitter(addr string, tcp bool, metricsPrefix string, logger log.Logger) Emitter {
	return &emitter{
		addr:   addr,
		tcp:    tcp,
		stop:   make(chan bool),
		prefix: metricsPrefix,
		logger: logger,
	}
}

// NewCounter returns a Counter whose value will be periodically emitted in
// a Graphite-compatible format once the Emitter is started. Fields are ignored.
func (e *emitter) NewCounter(name string) metrics.Counter {
	// only one flush at a time
	c := &counter{name, 0}
	e.mtx.Lock()
	e.counters = append(e.counters, c)
	e.mtx.Unlock()
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
func (e *emitter) NewHistogram(name string, minValue, maxValue int64, sigfigs int, quantiles ...int) (metrics.Histogram, error) {
	gauges := map[int]metrics.Gauge{}
	for _, quantile := range quantiles {
		if quantile <= 0 || quantile >= 100 {
			return nil, fmt.Errorf("invalid quantile %d", quantile)
		}
		gauges[quantile] = e.gauge(fmt.Sprintf("%s_p%02d", name, quantile))
	}
	h := newWindowedHistogram(name, minValue, maxValue, sigfigs, gauges, e.logger)

	e.mtx.Lock()
	e.histograms = append(e.histograms, h)
	e.mtx.Unlock()
	return h, nil
}

// NewTimeHistogram returns a TimeHistogram wrapper around the windowed
// HDR histrogram provided by this package.
func (e *emitter) NewTimeHistogram(name string, unit time.Duration, minValue, maxValue int64, sigfigs int, quantiles ...int) (metrics.TimeHistogram, error) {
	h, err := e.NewHistogram(name, minValue, maxValue, sigfigs, quantiles...)
	if err != nil {
		return nil, err
	}
	return metrics.NewTimeHistogram(unit, h), nil
}

// NewGauge returns a Gauge whose value will be periodically emitted in
// a Graphite-compatible format once the Emitter is started. Fields are ignored.
func (e *emitter) NewGauge(name string) metrics.Gauge {
	e.mtx.Lock()
	defer e.mtx.Unlock()
	return e.gauge(name)
}

func (e *emitter) gauge(name string) metrics.Gauge {
	g := &gauge{name, 0}
	e.gauges = append(e.gauges, g)
	return g
}

func (e *emitter) dial() error {
	if e.tcp {
		tAddr, err := net.ResolveTCPAddr("tcp", e.addr)
		if err != nil {
			return err
		}
		e.conn, err = net.DialTCP("tcp", nil, tAddr)
		if err != nil {
			return err
		}
	} else {
		uAddr, err := net.ResolveUDPAddr("udp", e.addr)
		if err != nil {
			return err
		}
		e.conn, err = net.DialUDP("udp", nil, uAddr)
		if err != nil {
			return err
		}
	}
	return nil
}

// Start will kick off a background goroutine to
// call Flush once every interval.
func (e *emitter) Start(interval time.Duration) error {
	var err error
	e.start.Do(func() {
		err = e.dial()
		if err != nil {
			return
		}
		go func() {
			t := time.Tick(interval)
			for {
				select {
				case <-t:
					e.Flush()
				case <-e.stop:
					return
				}
			}
		}()
	})
	return err
}

// Stop will flush the current metrics and close the
// current Graphite connection, if it exists.
func (e *emitter) Stop() error {
	if e.conn == nil {
		return nil
	}
	// stop the ticking flush loop
	e.stop <- true
	// get one last flush in
	e.Flush()
	// close the connection
	return e.conn.Close()
}

// Flush will attempt to create a connection with the given address
// and write the current metrics to it in the Graphite plaintext protocol.
//
// Users can call this method on process shutdown to ensure
// the current metrics are pushed to Graphite.
func (e *emitter) Flush() { e.flush(e.conn) }

func (e *emitter) flush(conn io.Writer) {
	// only one flush at a time
	e.mtx.Lock()
	defer e.mtx.Unlock()

	// buffer the writer and make sure to flush it
	w := bufio.NewWriter(conn)
	defer w.Flush()

	// emit counter stats
	for _, c := range e.counters {
		fmt.Fprintf(w, "%s.%s.count %d %d\n", e.prefix, c.Name(), c.count, time.Now().Unix())
	}

	// emit histogram specific stats
	for _, h := range e.histograms {
		hist := h.hist.Merge()
		now := time.Now().Unix()
		fmt.Fprintf(w, "%s.%s.count %d %d\n", e.prefix, h.Name(), hist.TotalCount(), now)
		fmt.Fprintf(w, "%s.%s.min %d %d\n", e.prefix, h.Name(), hist.Min(), now)
		fmt.Fprintf(w, "%s.%s.max %d %d\n", e.prefix, h.Name(), hist.Max(), now)
		fmt.Fprintf(w, "%s.%s.mean %.2f %d\n", e.prefix, h.Name(), hist.Mean(), now)
		fmt.Fprintf(w, "%s.%s.std-dev %.2f %d\n", e.prefix, h.Name(), hist.StdDev(), now)
	}

	// emit gauge stats (which can include some histogram quantiles)
	for _, g := range e.gauges {
		fmt.Fprintf(w, "%s.%s %.2f %d\n", e.prefix, g.Name(), g.Get(), time.Now().Unix())
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
	mtx  sync.Mutex
	hist *hdrhistogram.WindowedHistogram

	name   string
	gauges map[int]metrics.Gauge
	logger log.Logger
}

// newWindowedHistogram is taken from http://github.com/codahale/metrics. It returns a
// windowed HDR histogram which drops data older than five minutes.
//
// The histogram exposes metrics for each passed quantile as gauges. Users are expected
// to provide their own set of Gauges for quantiles to make this Histogram work across multiple
// metrics providers.
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

func (h *windowedHistogram) Name() string                         { return h.name }
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
