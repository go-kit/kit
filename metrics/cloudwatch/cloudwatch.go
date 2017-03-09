package cloudwatch

import (
	"sync"

	"time"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/generic"
)

// CloudWatch receives metrics observations and forwards them to CloudWatch.
// Create a CloudWatch object, use it to create metrics, and pass those metrics as
// dependencies to the components that will use them.
//
// To regularly report metrics to CloudWatch, use the WriteLoop helper method.
type CloudWatch struct {
	mtx        sync.RWMutex
	namespace  string
	svc        cloudwatchiface.CloudWatchAPI
	counters   map[string]*Counter
	gauges     map[string]*Gauge
	histograms map[string]*Histogram
	logger     log.Logger
}

// New returns a CloudWatch object that may be used to create metrics. Namespace is
// applied to all created metrics and maps to the CloudWatch namespace.
// Callers must ensure that regular calls to Send are performed, either manually or with one of the helper methods.
func New(namespace string, logger log.Logger, svc cloudwatchiface.CloudWatchAPI) *CloudWatch {
	return &CloudWatch{
		namespace:  namespace,
		svc:        svc,
		counters:   map[string]*Counter{},
		gauges:     map[string]*Gauge{},
		histograms: map[string]*Histogram{},
		logger:     logger,
	}
}

// NewCounter returns a counter. Observations are aggregated and emitted once
// per write invocation.
func (cw *CloudWatch) NewCounter(name string) *Counter {
	c := NewCounter(name)
	cw.mtx.Lock()
	cw.counters[name] = c
	cw.mtx.Unlock()
	return c
}

// NewGauge returns a gauge. Observations are aggregated and emitted once per
// write invocation.
func (cw *CloudWatch) NewGauge(name string) *Gauge {
	g := NewGauge(name)
	cw.mtx.Lock()
	cw.gauges[name] = g
	cw.mtx.Unlock()
	return g
}

// NewHistogram returns a histogram. Observations are aggregated and emitted as
// per-quantile gauges, once per write invocation. 50 is a good default value
// for buckets.
func (cw *CloudWatch) NewHistogram(name string, buckets int) *Histogram {
	h := NewHistogram(name, buckets)
	cw.mtx.Lock()
	cw.histograms[name] = h
	cw.mtx.Unlock()
	return h
}

// WriteLoop is a helper method that invokes Send every
// time the passed channel fires. This method blocks until the channel is
// closed, so clients probably want to run it in its own goroutine. For typical
// usage, create a time.Ticker and pass its C channel to this method.
func (cw *CloudWatch) WriteLoop(c <-chan time.Time) {
	for range c {
		if err := cw.Send(); err != nil {
			cw.logger.Log("during", "Send", "err", err)
		}
	}
}

// Send will fire an api request to CloudWatch with the latest stats for
// all metrics.
func (cw *CloudWatch) Send() error {
	cw.mtx.RLock()
	defer cw.mtx.RUnlock()
	now := time.Now()

	datums := []*cloudwatch.MetricDatum{}

	for name, c := range cw.counters {
		datums = append(datums, &cloudwatch.MetricDatum{
			MetricName: aws.String(name),
			Dimensions: makeDimensions(c.c.LabelValues()...),
			Value:      aws.Float64(c.c.Value()),
			Timestamp:  aws.Time(now),
		})
	}

	for name, g := range cw.gauges {
		datums = append(datums, &cloudwatch.MetricDatum{
			MetricName: aws.String(name),
			Dimensions: makeDimensions(g.g.LabelValues()...),
			Value:      aws.Float64(g.g.Value()),
			Timestamp:  aws.Time(now),
		})
	}

	for name, h := range cw.histograms {
		for _, p := range []struct {
			s string
			f float64
		}{
			{"50", 0.50},
			{"90", 0.90},
			{"95", 0.95},
			{"99", 0.99},
		} {
			datums = append(datums, &cloudwatch.MetricDatum{
				MetricName: aws.String(fmt.Sprintf("%s_%s", name, p.s)),
				Dimensions: makeDimensions(h.h.LabelValues()...),
				Value:      aws.Float64(h.h.Quantile(p.f)),
				Timestamp:  aws.Time(now),
			})
		}
	}

	_, err := cw.svc.PutMetricData(&cloudwatch.PutMetricDataInput{
		Namespace:  aws.String(cw.namespace),
		MetricData: datums,
	})
	return err
}

// Counter is a CloudWatch counter metric.
type Counter struct {
	c *generic.Counter
}

// NewCounter returns a new usable counter metric.
func NewCounter(name string) *Counter {
	return &Counter{
		c: generic.NewCounter(name),
	}
}

// With implements counter
func (c *Counter) With(labelValues ...string) metrics.Counter {
	c.c = c.c.With(labelValues...).(*generic.Counter)
	return c
}

// Add implements counter.
func (c *Counter) Add(delta float64) {
	c.c.Add(delta)
}

// Gauge is a CloudWatch gauge metric.
type Gauge struct {
	g *generic.Gauge
}

// NewGauge returns a new usable gauge metric
func NewGauge(name string) *Gauge {
	return &Gauge{
		g: generic.NewGauge(name),
	}
}

// With implements gauge
func (g *Gauge) With(labelValues ...string) metrics.Gauge {
	g.g = g.g.With(labelValues...).(*generic.Gauge)
	return g
}

// Set implements gauge
func (g *Gauge) Set(value float64) {
	g.g.Set(value)
}

// Add implements gauge
func (g *Gauge) Add(delta float64) {
	g.g.Add(delta)
}

// Histogram is a CloudWatch histogram metric
type Histogram struct {
	h *generic.Histogram
}

// NewHistogram returns a new usable histogram metric
func NewHistogram(name string, buckets int) *Histogram {
	return &Histogram{
		h: generic.NewHistogram(name, buckets),
	}
}

// With implements histogram
func (h *Histogram) With(labelValues ...string) metrics.Histogram {
	h.h = h.h.With(labelValues...).(*generic.Histogram)
	return h
}

// Observe implements histogram
func (h *Histogram) Observe(value float64) {
	h.h.Observe(value)
}

func makeDimensions(labelValues ...string) []*cloudwatch.Dimension {
	dimensions := make([]*cloudwatch.Dimension, len(labelValues)/2)
	for i, j := 0, 0; i < len(labelValues); i, j = i+2, j+1 {
		dimensions[j] = &cloudwatch.Dimension{
			Name:  aws.String(labelValues[i]),
			Value: aws.String(labelValues[i+1]),
		}
	}
	return dimensions
}
