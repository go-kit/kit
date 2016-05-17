// Package influxdb implements a InfluxDB backend for package metrics.
package influxdb

import (
	"time"

	"github.com/go-kit/kit/metrics"
	stdinflux "github.com/influxdata/influxdb/client/v2"
)

type counter struct {
	key    string
	tags   []metrics.Field
	fields []metrics.Field
	value  uint64
	bp     stdinflux.BatchPoints
}

// NewCounter returns a Counter that writes values in the reportInterval
// to the given InfluxDB client, utilizing batching
func NewCounter(client stdinflux.Client, bp stdinflux.BatchPoints, key string, tags []metrics.Field, reportInterval time.Duration) metrics.Counter {
	return NewCounterTick(client, bp, key, tags, time.Tick(reportInterval))
}

// NewCounterTick is the same as NewCounter, but allows the user to pass a own
// channel to trigger the write process to the client
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

// NewGauge creates a new gauge instance, reporting points in the defined reportInterval
func NewGauge(client stdinflux.Client, bp stdinflux.BatchPoints, key string, tags []metrics.Field, reportInterval time.Duration) metrics.Gauge {
	return NewGaugeTick(client, bp, key, tags, time.Tick(reportInterval))
}

// NewGaugeTick is the same as NewGauge with a ticker channel instead of a interval
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

func watch(client stdinflux.Client, bp stdinflux.BatchPoints, reportTicker <-chan time.Time) {
	for range reportTicker {
		client.Write(bp)
	}
}
