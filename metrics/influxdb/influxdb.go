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
	go c.watch(client, bp, reportTicker)
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

func (c *counter) watch(client stdinflux.Client, bp stdinflux.BatchPoints, reportTicker <-chan time.Time) {
	for range reportTicker {
		client.Write(bp)
	}
}
