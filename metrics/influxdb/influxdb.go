// Package influxdb implements a InfluxDB backend for package metrics.
package influxdb

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/codahale/hdrhistogram"
	stdinflux "github.com/influxdata/influxdb/client/v2"

	"github.com/go-kit/kit/metrics"
)

type counter struct {
	key    string
	tags   []metrics.Field
	fields []metrics.Field
	value  uint64
	bp     stdinflux.BatchPoints
}

// NewCounter returns a Counter that writes values in the reportInterval
// to the given InfluxDB client, utilizing batching.
func NewCounter(client stdinflux.Client, bp stdinflux.BatchPoints, key string, tags []metrics.Field, reportInterval time.Duration) metrics.Counter {
	return NewCounterTick(client, bp, key, tags, time.Tick(reportInterval))
}

// NewCounterTick is the same as NewCounter, but allows the user to pass a own
// channel to trigger the write process to the client.
func NewCounterTick(client stdinflux.Client, bp stdinflux.BatchPoints, key string, tags []metrics.Field, reportTicker <-chan time.Time) metrics.Counter {
	c := &counter{
		key:   key,
		tags:  tags,
		value: 0,
		bp:    bp,
	}
	go watch(client, bp, reportTicker)
	return c
}

func (c *counter) Name() string {
	return c.key
}

func (c *counter) With(field metrics.Field) metrics.Counter {
	return &counter{
		key:    c.key,
		tags:   c.tags,
		value:  c.value,
		bp:     c.bp,
		fields: append(c.fields, field),
	}
}

func (c *counter) Add(delta uint64) {
	c.value = c.value + delta

	tags := map[string]string{}

	for _, tag := range c.tags {
		tags[tag.Key] = tag.Value
	}

	fields := map[string]interface{}{}

	for _, field := range c.fields {
		fields[field.Key] = field.Value
	}
	fields["value"] = c.value
	pt, _ := stdinflux.NewPoint(c.key, tags, fields, time.Now())
	c.bp.AddPoint(pt)
}

type gauge struct {
	key    string
	tags   []metrics.Field
	fields []metrics.Field
	value  float64
	bp     stdinflux.BatchPoints
}

// NewGauge creates a new gauge instance, reporting points in the defined reportInterval.
func NewGauge(client stdinflux.Client, bp stdinflux.BatchPoints, key string, tags []metrics.Field, reportInterval time.Duration) metrics.Gauge {
	return NewGaugeTick(client, bp, key, tags, time.Tick(reportInterval))
}

// NewGaugeTick is the same as NewGauge with a ticker channel instead of a interval.
func NewGaugeTick(client stdinflux.Client, bp stdinflux.BatchPoints, key string, tags []metrics.Field, reportTicker <-chan time.Time) metrics.Gauge {
	g := &gauge{
		key:   key,
		tags:  tags,
		value: 0,
		bp:    bp,
	}
	go watch(client, bp, reportTicker)
	return g
}

func (g *gauge) Name() string {
	return g.key
}

func (g *gauge) With(field metrics.Field) metrics.Gauge {
	return &gauge{
		key:    g.key,
		tags:   g.tags,
		value:  g.value,
		bp:     g.bp,
		fields: append(g.fields, field),
	}
}

func (g *gauge) Add(delta float64) {
	g.value = g.value + delta
	g.createPoint()
}

func (g *gauge) Set(value float64) {
	g.value = value
	g.createPoint()
}

func (g *gauge) Get() float64 {
	return g.value
}

func (g *gauge) createPoint() {
	tags := map[string]string{}

	for _, tag := range g.tags {
		tags[tag.Key] = tag.Value
	}

	fields := map[string]interface{}{}

	for _, field := range g.fields {
		fields[field.Key] = field.Value
	}
	fields["value"] = g.value
	pt, _ := stdinflux.NewPoint(g.key, tags, fields, time.Now())
	g.bp.AddPoint(pt)
}

// The implementation from histogram is taken from metrics/expvar

type histogram struct {
	mu   sync.Mutex
	hist *hdrhistogram.WindowedHistogram

	key    string
	gauges map[int]metrics.Gauge
}

// NewHistogram is taken from http://github.com/codahale/metrics. It returns a
// windowed HDR histogram which drops data older than five minutes.
//
// The histogram exposes metrics for each passed quantile as gauges. Quantiles
// should be integers in the range 1..99. The gauge names are assigned by
// using the passed name as a prefix and appending "_pNN" e.g. "_p50".
func NewHistogram(client stdinflux.Client, bp stdinflux.BatchPoints, key string, tags []metrics.Field,
	reportInterval time.Duration, minValue, maxValue int64, sigfigs int, quantiles ...int) metrics.Histogram {
	return NewHistogramTick(client, bp, key, tags, time.Tick(reportInterval), minValue, maxValue, sigfigs, quantiles...)
}

// NewHistogramTick is the same as NewHistoGram, but allows to pass a custom reportTicker.
func NewHistogramTick(client stdinflux.Client, bp stdinflux.BatchPoints, key string, tags []metrics.Field,
	reportTicker <-chan time.Time, minValue, maxValue int64, sigfigs int, quantiles ...int) metrics.Histogram {
	gauges := map[int]metrics.Gauge{}

	for _, quantile := range quantiles {
		if quantile <= 0 || quantile >= 100 {
			panic(fmt.Sprintf("invalid quantile %d", quantile))
		}
		gauges[quantile] = NewGaugeTick(client, bp, fmt.Sprintf("%s_p%02d", key, quantile), tags, reportTicker)
	}

	h := &histogram{
		hist:   hdrhistogram.NewWindowed(5, minValue, maxValue, sigfigs),
		key:    key,
		gauges: gauges,
	}

	go h.rotateLoop(1 * time.Minute)
	return h
}

func (h *histogram) Name() string {
	return h.key
}

func (h *histogram) With(field metrics.Field) metrics.Histogram {
	for q, gauge := range h.gauges {
		h.gauges[q] = gauge.With(field)
	}

	return h
}

func (h *histogram) Observe(value int64) {
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

func (h *histogram) Distribution() ([]metrics.Bucket, []metrics.Quantile) {
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

func (h *histogram) rotateLoop(d time.Duration) {
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

func watch(client stdinflux.Client, bp stdinflux.BatchPoints, reportTicker <-chan time.Time) {
	for range reportTicker {
		client.Write(bp)
	}
}
