// Package statsd implements a statsd backend for package metrics.
//
// The current implementation ignores fields. In the future, it would be good
// to have an implementation that accepted a set of predeclared field names at
// construction time, and used field values to produce delimiter-separated
// bucket (key) names. That is,
//
//    c := NewFieldedCounter(..., "path", "status")
//    c.Add(1) // "myprefix.unknown.unknown:1|c\n"
//    c2 := c.With("path", "foo").With("status": "200")
//    c2.Add(1) // "myprefix.foo.200:1|c\n"
//
package statsd

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math"
	"time"

	"sync/atomic"

	"github.com/go-kit/kit/metrics"
)

// statsd metrics take considerable influence from
// https://github.com/streadway/handy package statsd.

const maxBufferSize = 1400 // bytes

type statsdCounter struct {
	key string
	c   chan string
}

// NewCounter returns a Counter that emits observations in the statsd protocol
// to the passed writer. Observations are buffered for the report interval or
// until the buffer exceeds a max packet size, whichever comes first. Fields
// are ignored.
//
// TODO: support for sampling.
func NewCounter(w io.Writer, key string, reportInterval time.Duration) metrics.Counter {
	return NewCounterTick(w, key, time.Tick(reportInterval))
}

// NewCounterTick is the same as NewCounter, but allows the user to pass in a
// ticker channel instead of invoking time.Tick.
func NewCounterTick(w io.Writer, key string, reportTicker <-chan time.Time) metrics.Counter {
	c := &statsdCounter{
		key: key,
		c:   make(chan string),
	}
	go fwd(w, key, reportTicker, c.c)
	return c
}

func (c *statsdCounter) Name() string { return c.key }

func (c *statsdCounter) With(metrics.Field) metrics.Counter { return c }

func (c *statsdCounter) Add(delta uint64) { c.c <- fmt.Sprintf("%d|c", delta) }

type statsdGauge struct {
	key       string
	lastValue uint64 // math.Float64frombits
	g         chan string
}

// NewGauge returns a Gauge that emits values in the statsd protocol to the
// passed writer. Values are buffered for the report interval or until the
// buffer exceeds a max packet size, whichever comes first. Fields are
// ignored.
//
// TODO: support for sampling.
func NewGauge(w io.Writer, key string, reportInterval time.Duration) metrics.Gauge {
	return NewGaugeTick(w, key, time.Tick(reportInterval))
}

// NewGaugeTick is the same as NewGauge, but allows the user to pass in a ticker
// channel instead of invoking time.Tick.
func NewGaugeTick(w io.Writer, key string, reportTicker <-chan time.Time) metrics.Gauge {
	g := &statsdGauge{
		key: key,
		g:   make(chan string),
	}
	go fwd(w, key, reportTicker, g.g)
	return g
}

func (g *statsdGauge) Name() string { return g.key }

func (g *statsdGauge) With(metrics.Field) metrics.Gauge { return g }

func (g *statsdGauge) Add(delta float64) {
	// https://github.com/etsy/statsd/blob/master/docs/metric_types.md#gauges
	sign := "+"
	if delta < 0 {
		sign, delta = "-", -delta
	}
	g.g <- fmt.Sprintf("%s%f|g", sign, delta)
}

func (g *statsdGauge) Set(value float64) {
	atomic.StoreUint64(&g.lastValue, math.Float64bits(value))
	g.g <- fmt.Sprintf("%f|g", value)
}

func (g *statsdGauge) Get() float64 {
	return math.Float64frombits(atomic.LoadUint64(&g.lastValue))
}

// NewCallbackGauge emits values in the statsd protocol to the passed writer.
// It collects values every scrape interval from the callback. Values are
// buffered for the report interval or until the buffer exceeds a max packet
// size, whichever comes first. The report and scrape intervals may be the
// same. The callback determines the value, and fields are ignored, so
// NewCallbackGauge returns nothing.
func NewCallbackGauge(w io.Writer, key string, reportInterval, scrapeInterval time.Duration, callback func() float64) {
	NewCallbackGaugeTick(w, key, time.Tick(reportInterval), time.Tick(scrapeInterval), callback)
}

// NewCallbackGaugeTick is the same as NewCallbackGauge, but allows the user to
// pass in ticker channels instead of durations to control report and scrape
// intervals.
func NewCallbackGaugeTick(w io.Writer, key string, reportTicker, scrapeTicker <-chan time.Time, callback func() float64) {
	go fwd(w, key, reportTicker, emitEvery(scrapeTicker, callback))
}

func emitEvery(emitTicker <-chan time.Time, callback func() float64) <-chan string {
	c := make(chan string)
	go func() {
		for range emitTicker {
			c <- fmt.Sprintf("%f|g", callback())
		}
	}()
	return c
}

type statsdHistogram struct {
	key string
	h   chan string
}

// NewHistogram returns a Histogram that emits observations in the statsd
// protocol to the passed writer. Observations are buffered for the reporting
// interval or until the buffer exceeds a max packet size, whichever comes
// first. Fields are ignored.
//
// NewHistogram is mapped to a statsd Timing, so observations should represent
// milliseconds. If you observe in units of nanoseconds, you can make the
// translation with a ScaledHistogram:
//
//    NewScaledHistogram(statsdHistogram, time.Millisecond)
//
// You can also enforce the constraint in a typesafe way with a millisecond
// TimeHistogram:
//
//    NewTimeHistogram(statsdHistogram, time.Millisecond)
//
// TODO: support for sampling.
func NewHistogram(w io.Writer, key string, reportInterval time.Duration) metrics.Histogram {
	return NewHistogramTick(w, key, time.Tick(reportInterval))
}

// NewHistogramTick is the same as NewHistogram, but allows the user to pass a
// ticker channel instead of invoking time.Tick.
func NewHistogramTick(w io.Writer, key string, reportTicker <-chan time.Time) metrics.Histogram {
	h := &statsdHistogram{
		key: key,
		h:   make(chan string),
	}
	go fwd(w, key, reportTicker, h.h)
	return h
}

func (h *statsdHistogram) Name() string { return h.key }

func (h *statsdHistogram) With(metrics.Field) metrics.Histogram { return h }

func (h *statsdHistogram) Observe(value int64) {
	h.h <- fmt.Sprintf("%d|ms", value)
}

func (h *statsdHistogram) Distribution() ([]metrics.Bucket, []metrics.Quantile) {
	// TODO(pb): no way to do this without introducing e.g. codahale/hdrhistogram
	return []metrics.Bucket{}, []metrics.Quantile{}
}

func fwd(w io.Writer, key string, reportTicker <-chan time.Time, c <-chan string) {
	buf := &bytes.Buffer{}
	for {
		select {
		case s := <-c:
			fmt.Fprintf(buf, "%s:%s\n", key, s)
			if buf.Len() > maxBufferSize {
				flush(w, buf)
			}

		case <-reportTicker:
			flush(w, buf)
		}
	}
}

func flush(w io.Writer, buf *bytes.Buffer) {
	if buf.Len() <= 0 {
		return
	}
	if _, err := w.Write(buf.Bytes()); err != nil {
		log.Printf("error: could not write to statsd: %v", err)
	}
	buf.Reset()
}
