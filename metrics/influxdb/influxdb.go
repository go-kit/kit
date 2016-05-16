// Package influxdb implements a Influxdb backend for package metrics.
package influxdb

import (
	"time"

	"github.com/go-kit/kit/metrics"
	stdinflux "github.com/influxdata/influxdb/client/v2"
)

type influxdbCounter struct {
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
	counter := &influxdbCounter{
		key:   key,
		tags:  tags,
		value: 0,
		bp:    bp,
	}
	go counter.watch(client, bp, reportTicker)
	return counter
}

func (counter *influxdbCounter) Name() string {
	return counter.key
}

func (counter *influxdbCounter) With(field metrics.Field) metrics.Counter {
	return &influxdbCounter{
		key:    counter.key,
		tags:   counter.tags,
		value:  counter.value,
		bp:     counter.bp,
		fields: append(counter.fields, field),
	}
}

func (counter *influxdbCounter) Add(delta uint64) {
	counter.value = counter.value + delta

	tags := map[string]string{}

	for _, tag := range counter.tags {
		tags[tag.Key] = tag.Value
	}

	fields := map[string]interface{}{}

	for _, field := range counter.fields {
		fields[field.Key] = field.Value
	}
	fields["value"] = counter.value
	pt, _ := stdinflux.NewPoint(counter.key, tags, fields, time.Now())
	counter.bp.AddPoint(pt)
}

func (counter influxdbCounter) watch(client stdinflux.Client, bp stdinflux.BatchPoints, reportTicker <-chan time.Time) {
	for {
		select {
		case <-reportTicker:
			client.Write(bp)
		}
	}
}
