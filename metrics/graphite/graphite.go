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
	"fmt"
	"io"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/codahale/hdrhistogram"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
)

func newCounter(name string) *counter {
	return &counter{name, 0}
}

func newGauge(name string) *gauge {
	return &gauge{name, 0}
}

// counter implements the metrics.counter interface but also provides a
// Flush method to emit the current counter values in the Graphite plaintext
// protocol.
type counter struct {
	key   string
	count uint64
}

func (c *counter) Name() string { return c.key }

// With currently ignores fields.
func (c *counter) With(metrics.Field) metrics.Counter { return c }

func (c *counter) Add(delta uint64) { atomic.AddUint64(&c.count, delta) }

func (c *counter) get() uint64 { return atomic.LoadUint64(&c.count) }

// flush will emit the current counter value in the Graphite plaintext
// protocol to the given io.Writer.
func (c *counter) flush(w io.Writer, prefix string) {
	fmt.Fprintf(w, "%s.count %d %d\n", prefix+c.Name(), c.get(), time.Now().Unix())
}

// gauge implements the metrics.gauge interface but also provides a
// Flush method to emit the current counter values in the Graphite plaintext
// protocol.
type gauge struct {
	key   string
	value uint64 // math.Float64bits
}

func (g *gauge) Name() string { return g.key }

// With currently ignores fields.
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

// Flush will emit the current gauge value in the Graphite plaintext
// protocol to the given io.Writer.
func (g *gauge) flush(w io.Writer, prefix string) {
	fmt.Fprintf(w, "%s %.2f %d\n", prefix+g.Name(), g.Get(), time.Now().Unix())
}

// windowedHistogram is taken from http://github.com/codahale/metrics. It
// is a windowed HDR histogram which drops data older than five minutes.
//
// The histogram exposes metrics for each passed quantile as gauges. Quantiles
// should be integers in the range 1..99. The gauge names are assigned by using
// the passed name as a prefix and appending "_pNN" e.g. "_p50".
//
// The values of this histogram will be periodically emitted in a
// Graphite-compatible format once the GraphiteProvider is started. Fields are ignored.
type windowedHistogram struct {
	mtx  sync.Mutex
	hist *hdrhistogram.WindowedHistogram

	name   string
	gauges map[int]metrics.Gauge
	logger log.Logger
}

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

func (h *windowedHistogram) flush(w io.Writer, prefix string) {
	name := prefix + h.Name()
	hist := h.hist.Merge()
	now := time.Now().Unix()
	fmt.Fprintf(w, "%s.count %d %d\n", name, hist.TotalCount(), now)
	fmt.Fprintf(w, "%s.min %d %d\n", name, hist.Min(), now)
	fmt.Fprintf(w, "%s.max %d %d\n", name, hist.Max(), now)
	fmt.Fprintf(w, "%s.mean %.2f %d\n", name, hist.Mean(), now)
	fmt.Fprintf(w, "%s.std-dev %.2f %d\n", name, hist.StdDev(), now)
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
