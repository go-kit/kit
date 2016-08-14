// Package influx provides an InfluxDB implementation for metrics. The model is
// similar to other push-based instrumentation systems. Observations are
// aggregated locally and emitted to the Influx server on regular intervals.
package influx

import (
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics2"
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

// New returns an Influx, ready to create metrics and collect observations. Tags
// are applied to all metrics created from this object. The BatchPointsConfig is
// used during flushing.
func New(tags map[string]string, conf influxdb.BatchPointsConfig, logger log.Logger) *Influx {
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

// WriteLoop is a helper method that invokes WriteTo to the passed writer every
// time the passed channel fires. This method blocks until the channel is
// closed, so clients probably want to run it in its own goroutine. For typical
// usage, create a time.Ticker and pass its C channel to this method.
func (i *Influx) WriteLoop(c <-chan time.Time, w BatchPointsWriter) {
	for range c {
		if err := i.WriteTo(w); err != nil {
			i.logger.Log("during", "WriteTo", "err", err)
		}
	}
}

// BatchPointsWriter captures a subset of the influxdb.Client methods necessary
// for emitting metrics observations.
type BatchPointsWriter interface {
	Write(influxdb.BatchPoints) error
}

// WriteTo converts the current set of metrics to Influx BatchPoints, and writes
// the BatchPoints to the client. Clients probably shouldn't invoke this method
// directly, and should prefer using WriteLoop.
func (i *Influx) WriteTo(w BatchPointsWriter) error {
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
		p, err := influxdb.NewPoint(name, merge(i.tags, c.tags), fields, now)
		if err != nil {
			return err
		}
		bp.AddPoint(p)
	}

	for name, g := range i.gauges {
		fields := fieldsFrom(g.LabelValues())
		fields["value"] = g.Value()
		p, err := influxdb.NewPoint(name, merge(i.tags, g.tags), fields, now)
		if err != nil {
			return err
		}
		bp.AddPoint(p)
	}

	for name, h := range i.histograms {
		fields := fieldsFrom(h.LabelValues())
		for _, x := range []struct {
			suffix   string
			quantile float64
		}{
			{".p50", 0.50},
			{".p90", 0.90},
			{".p95", 0.95},
			{".p99", 0.99},
		} {
			fields["value"] = h.Quantile(x.quantile)
			p, err := influxdb.NewPoint(name+x.suffix, merge(i.tags, h.tags), fields, now)
			if err != nil {
				return err
			}
			bp.AddPoint(p)
		}
	}

	return w.Write(bp)
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

func merge(a, b map[string]string) map[string]string {
	res := map[string]string{}
	for k, v := range a {
		res[k] = v
	}
	for k, v := range b {
		res[k] = v
	}
	return res
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

// With adapts the generic With method to update the metric in-place. This is
// necessary so that pointers in the parent Influx struct aren't invalidated.
func (c *Counter) With(labelValues ...string) metrics.Counter {
	c.Counter = c.Counter.With(labelValues...).(*generic.Counter)
	return c
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

// With adapts the generic With method to update the metric in-place. This is
// necessary so that pointers in the parent Influx struct aren't invalidated.
func (g *Gauge) With(labelValues ...string) metrics.Gauge {
	g.Gauge = g.Gauge.With(labelValues...).(*generic.Gauge)
	return g
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

// With adapts the generic With method to update the metric in-place. This is
// necessary so that pointers in the parent Influx struct aren't invalidated.
func (h *Histogram) With(labelValues ...string) metrics.Histogram {
	h.Histogram = h.Histogram.With(labelValues...).(*generic.Histogram)
	return h
}
