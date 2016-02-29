// Package dogstatsd implements a DogStatsD backend for package metrics.
//
// This implementation supports Datadog tags that provide additional metric
// filtering capabilities. See the DogStatsD documentation for protocol
// specifics:
// http://docs.datadoghq.com/guides/dogstatsd/
//
package dogstatsd

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

// dogstatsd metrics were based on the statsd package in go-kit

const maxBufferSize = 1400 // bytes

type dogstatsdCounter struct {
	key  string
	c    chan string
	tags []metrics.Field
}

// NewCounter returns a Counter that emits observations in the DogStatsD protocol
// to the passed writer. Observations are buffered for the report interval or
// until the buffer exceeds a max packet size, whichever comes first.
//
// TODO: support for sampling.
func NewCounter(w io.Writer, key string, reportInterval time.Duration, globalTags []metrics.Field) metrics.Counter {
	return NewCounterTick(w, key, time.Tick(reportInterval), globalTags)
}

// NewCounterTick is the same as NewCounter, but allows the user to pass in a
// ticker channel instead of invoking time.Tick.
func NewCounterTick(w io.Writer, key string, reportTicker <-chan time.Time, tags []metrics.Field) metrics.Counter {
	c := &dogstatsdCounter{
		key:  key,
		c:    make(chan string),
		tags: tags,
	}
	go fwd(w, key, reportTicker, c.c)
	return c
}

func (c *dogstatsdCounter) Name() string { return c.key }

func (c *dogstatsdCounter) With(f metrics.Field) metrics.Counter {
	return &dogstatsdCounter{
		key:  c.key,
		c:    c.c,
		tags: append(c.tags, f),
	}
}

func (c *dogstatsdCounter) Add(delta uint64) { c.c <- applyTags(fmt.Sprintf("%d|c", delta), c.tags) }

type dogstatsdGauge struct {
	key       string
	lastValue uint64 // math.Float64frombits
	g         chan string
	tags      []metrics.Field
}

// NewGauge returns a Gauge that emits values in the DogStatsD protocol to the
// passed writer. Values are buffered for the report interval or until the
// buffer exceeds a max packet size, whichever comes first.
//
// TODO: support for sampling.
func NewGauge(w io.Writer, key string, reportInterval time.Duration, tags []metrics.Field) metrics.Gauge {
	return NewGaugeTick(w, key, time.Tick(reportInterval), tags)
}

// NewGaugeTick is the same as NewGauge, but allows the user to pass in a ticker
// channel instead of invoking time.Tick.
func NewGaugeTick(w io.Writer, key string, reportTicker <-chan time.Time, tags []metrics.Field) metrics.Gauge {
	g := &dogstatsdGauge{
		key:  key,
		g:    make(chan string),
		tags: tags,
	}
	go fwd(w, key, reportTicker, g.g)
	return g
}

func (g *dogstatsdGauge) Name() string { return g.key }

func (g *dogstatsdGauge) With(f metrics.Field) metrics.Gauge {
	return &dogstatsdGauge{
		key:       g.key,
		lastValue: g.lastValue,
		g:         g.g,
		tags:      append(g.tags, f),
	}
}

func (g *dogstatsdGauge) Add(delta float64) {
	// https://github.com/etsy/statsd/blob/master/docs/metric_types.md#gauges
	sign := "+"
	if delta < 0 {
		sign, delta = "-", -delta
	}
	g.g <- applyTags(fmt.Sprintf("%s%f|g", sign, delta), g.tags)
}

func (g *dogstatsdGauge) Set(value float64) {
	atomic.StoreUint64(&g.lastValue, math.Float64bits(value))
	g.g <- applyTags(fmt.Sprintf("%f|g", value), g.tags)
}

func (g *dogstatsdGauge) Get() float64 {
	return math.Float64frombits(atomic.LoadUint64(&g.lastValue))
}

// NewCallbackGauge emits values in the DogStatsD protocol to the passed writer.
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

type dogstatsdHistogram struct {
	key  string
	h    chan string
	tags []metrics.Field
}

// NewHistogram returns a Histogram that emits observations in the DogStatsD
// protocol to the passed writer. Observations are buffered for the reporting
// interval or until the buffer exceeds a max packet size, whichever comes
// first.
//
// NewHistogram is mapped to a statsd Timing, so observations should represent
// milliseconds. If you observe in units of nanoseconds, you can make the
// translation with a ScaledHistogram:
//
//    NewScaledHistogram(dogstatsdHistogram, time.Millisecond)
//
// You can also enforce the constraint in a typesafe way with a millisecond
// TimeHistogram:
//
//    NewTimeHistogram(dogstatsdHistogram, time.Millisecond)
//
// TODO: support for sampling.
func NewHistogram(w io.Writer, key string, reportInterval time.Duration, tags []metrics.Field) metrics.Histogram {
	return NewHistogramTick(w, key, time.Tick(reportInterval), tags)
}

// NewHistogramTick is the same as NewHistogram, but allows the user to pass a
// ticker channel instead of invoking time.Tick.
func NewHistogramTick(w io.Writer, key string, reportTicker <-chan time.Time, tags []metrics.Field) metrics.Histogram {
	h := &dogstatsdHistogram{
		key:  key,
		h:    make(chan string),
		tags: tags,
	}
	go fwd(w, key, reportTicker, h.h)
	return h
}

func (h *dogstatsdHistogram) Name() string { return h.key }

func (h *dogstatsdHistogram) With(f metrics.Field) metrics.Histogram {
	return &dogstatsdHistogram{
		key:  h.key,
		h:    h.h,
		tags: append(h.tags, f),
	}
}

func (h *dogstatsdHistogram) Observe(value int64) {
	h.h <- applyTags(fmt.Sprintf("%d|ms", value), h.tags)
}

func (h *dogstatsdHistogram) Distribution() ([]metrics.Bucket, []metrics.Quantile) {
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
		log.Printf("error: could not write to dogstatsd: %v", err)
	}
	buf.Reset()
}

func applyTags(value string, tags []metrics.Field) string {
	if len(tags) > 0 {
		var tagsString string
		for _, t := range tags {
			switch tagsString {
			case "":
				tagsString = t.Key + ":" + t.Value
			default:
				tagsString = tagsString + "," + t.Key + ":" + t.Value
			}
		}
		value = value + "|#" + tagsString
	}
	return value
}
