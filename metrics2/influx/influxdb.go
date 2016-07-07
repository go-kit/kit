// Package influx provides an InfluxDB implementation for metrics. The model is
// similar to other push-based instrumentation systems. Observations are
// aggregated locally and emitted to the Influx server on regular intervals.
package influx

import (
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics2/generic"
	influxdb "github.com/influxdata/influxdb/client/v2"
)

// Influx is a store for metrics that will be emitted to an Influx database.
//
// Influx is a general purpose time-series database, and has no native concepts
// of counters, gauges, or histograms. Counters are modeled as a timeseries with
// one data point per flush, with a "count" field that reflects all adds since
// the last flush. Gauges are modeled as a timeseries with one data point per
// flush, with a "value" field that reflects the current state of the gauge.
// Histograms are modeled as 4 gauge timeseries, one each for the 50th, 90th,
// 95th, and 99th quantiles.
//
// Influx tags are assigned to each Go kit metric at construction, and are
// immutable for the life of the metric. Influx fields are mapped to Go kit
// label values, and may be mutated via With functions. Actual metric values are
// provided as fields with specific names depending on the metric.
//
// All observations are batched in memory locally, and flushed on demand.
type Influx struct {
	mtx        sync.RWMutex
	counters   map[string]*Counter
	gauges     map[string]*Gauge
	histograms map[string]*Histogram
	tags       map[string]string
	conf       influxdb.BatchPointsConfig
	logger     log.Logger
}

// New returns an Influx object, ready to create metrics and aggregate
// observations, and automatically flushing to the passed Influx client every
// flushInterval. Use the returned stop function to terminate the flushing
// goroutine.
func New(
	tags map[string]string,
	conf influxdb.BatchPointsConfig,
	client influxdb.Client,
	flushInterval time.Duration,
	logger log.Logger,
) (res *Influx, stop func()) {
	i := NewRaw(tags, conf, logger)
	ticker := time.NewTicker(flushInterval)
	go i.FlushTo(client, ticker)
	return i, ticker.Stop
}

// NewRaw returns an Influx object, ready to create metrics and aggregate
// observations, but without automatically flushing anywhere. Users should
// probably prefer the New constructor.
//
// Tags are applied to all metrics created from this object. A BatchPoints
// structure is created from the provided BatchPointsConfig; any error will
// cause a panic. Observations are aggregated into the BatchPoints.
func NewRaw(tags map[string]string, conf influxdb.BatchPointsConfig, logger log.Logger) *Influx {
	return &Influx{
		counters:   map[string]*Counter{},
		gauges:     map[string]*Gauge{},
		histograms: map[string]*Histogram{},
		tags:       tags,
		conf:       conf,
		logger:     logger,
	}
}

// NewCounter returns a generic counter with static tags.
func (i *Influx) NewCounter(name string, tags map[string]string) *Counter {
	i.mtx.Lock()
	defer i.mtx.Unlock()
	c := newCounter(tags)
	i.counters[name] = c
	return c
}

// NewGauge returns a generic gauge with static tags.
func (i *Influx) NewGauge(name string, tags map[string]string) *Gauge {
	i.mtx.Lock()
	defer i.mtx.Unlock()
	g := newGauge(tags)
	i.gauges[name] = g
	return g
}

// NewHistogram returns a generic histogram with static tags. 50 is a good
// default number of buckets.
func (i *Influx) NewHistogram(name string, tags map[string]string, buckets int) *Histogram {
	i.mtx.Lock()
	defer i.mtx.Unlock()
	h := newHistogram(tags, buckets)
	i.histograms[name] = h
	return h
}

// FlushTo invokes WriteTo to the client every time the ticker fires. FlushTo
// blocks until the ticker is stopped. Most users won't need to call this method
// directly, and should prefer to use the New constructor.
func (i *Influx) FlushTo(client influxdb.Client, ticker *time.Ticker) {
	for range ticker.C {
		if err := i.WriteTo(client); err != nil {
			i.logger.Log("during", "flush", "err", err)
		}
	}
}

// WriteTo converts the current set of metrics to Influx BatchPoints, and writes
// the BatchPoints to the client. Clients probably shouldn't invoke this method
// directly, and should prefer using FlushTo, or the New constructor.
func (i *Influx) WriteTo(client influxdb.Client) error {
	i.mtx.Lock()
	defer i.mtx.Unlock()

	bp, err := influxdb.NewBatchPoints(i.conf)
	if err != nil {
		return err
	}
	now := time.Now()

	for name, c := range i.counters {
		fields := fieldsFrom(c.LabelValues())
		fields["count"] = c.ValueReset()
		p, err := influxdb.NewPoint(name, c.tags, fields, now)
		if err != nil {
			return err
		}
		bp.AddPoint(p)
	}

	for name, g := range i.gauges {
		fields := fieldsFrom(g.LabelValues())
		fields["value"] = g.Value()
		p, err := influxdb.NewPoint(name, g.tags, fields, now)
		if err != nil {
			return err
		}
		bp.AddPoint(p)
	}

	for name, h := range i.histograms {
		fields := fieldsFrom(h.LabelValues())
		for suffix, quantile := range map[string]float64{
			".p50": 0.50,
			".p90": 0.90,
			".p95": 0.95,
			".p99": 0.99,
		} {
			fields["value"] = h.Quantile(quantile)
			p, err := influxdb.NewPoint(name+suffix, h.tags, fields, now)
			if err != nil {
				return err
			}
			bp.AddPoint(p)
		}
	}

	return client.Write(bp)
}

func fieldsFrom(labelValues []string) map[string]interface{} {
	if len(labelValues)%2 != 0 {
		panic("fieldsFrom received a labelValues with an odd number of strings")
	}
	fields := make(map[string]interface{}, len(labelValues)/2)
	for i := 0; i < len(labelValues); i += 2 {
		fields[labelValues[i]] = labelValues[i+1]
	}
	return fields
}

// Counter is a generic counter, with static tags.
type Counter struct {
	*generic.Counter
	tags map[string]string
}

func newCounter(tags map[string]string) *Counter {
	return &Counter{
		Counter: generic.NewCounter(),
		tags:    tags,
	}
}

// Gauge is a generic gauge, with static tags.
type Gauge struct {
	*generic.Gauge
	tags map[string]string
}

func newGauge(tags map[string]string) *Gauge {
	return &Gauge{
		Gauge: generic.NewGauge(),
		tags:  tags,
	}
}

// Histogram is a generic histogram, with static tags.
type Histogram struct {
	*generic.Histogram
	tags map[string]string
}

func newHistogram(tags map[string]string, buckets int) *Histogram {
	return &Histogram{
		Histogram: generic.NewHistogram(buckets),
		tags:      tags,
	}
}
