package cloudwatch

import (
	"fmt"
	"sync"
	"time"

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
	counters   map[string]*counter
	gauges     map[string]*gauge
	histograms map[string]*histogram
	logger     log.Logger
}

// New returns a CloudWatch object that may be used to create metrics.
// Namespace is applied to all created metrics and maps to the CloudWatch namespace.
// Callers must ensure that regular calls to Send are performed, either
// manually or with one of the helper methods.
func New(namespace string, svc cloudwatchiface.CloudWatchAPI, logger log.Logger) *CloudWatch {
	return &CloudWatch{
		namespace:  namespace,
		svc:        svc,
		counters:   map[string]*counter{},
		gauges:     map[string]*gauge{},
		histograms: map[string]*histogram{},
		logger:     logger,
	}
}

// NewCounter returns a counter. Observations are aggregated and emitted once
// per write invocation.
func (cw *CloudWatch) NewCounter(name string) metrics.Counter {
	cw.mtx.Lock()
	defer cw.mtx.Unlock()
	c := &counter{c: generic.NewCounter(name)}
	cw.counters[name] = c
	return c
}

// NewGauge returns a gauge. Observations are aggregated and emitted once per
// write invocation.
func (cw *CloudWatch) NewGauge(name string) metrics.Gauge {
	cw.mtx.Lock()
	defer cw.mtx.Unlock()
	g := &gauge{g: generic.NewGauge(name)}
	cw.gauges[name] = g
	return g
}

// NewHistogram returns a histogram. Observations are aggregated and emitted as
// per-quantile gauges, once per write invocation. 50 is a good default value
// for buckets.
func (cw *CloudWatch) NewHistogram(name string, buckets int) metrics.Histogram {
	cw.mtx.Lock()
	defer cw.mtx.Unlock()
	h := &histogram{h: generic.NewHistogram(name, buckets)}
	cw.histograms[name] = h
	return h
}

// WriteLoop is a helper method that invokes Send every time the passed
// channel fires. This method blocks until the channel is closed, so clients
// probably want to run it in its own goroutine. For typical usage, create a
// time.Ticker and pass its C channel to this method.
func (cw *CloudWatch) WriteLoop(c <-chan time.Time) {
	for range c {
		if err := cw.Send(); err != nil {
			cw.logger.Log("during", "Send", "err", err)
		}
	}
}

// Send will fire an API request to CloudWatch with the latest stats for
// all metrics. It is preferred that the WriteLoop method is used.
func (cw *CloudWatch) Send() error {
	cw.mtx.RLock()
	defer cw.mtx.RUnlock()
	now := time.Now()

	var datums []*cloudwatch.MetricDatum

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

// counter is a CloudWatch counter metric.
type counter struct {
	c *generic.Counter
}

// With implements counter
func (c *counter) With(labelValues ...string) metrics.Counter {
	c.c = c.c.With(labelValues...).(*generic.Counter)
	return c
}

// Add implements counter.
func (c *counter) Add(delta float64) {
	c.c.Add(delta)
}

// gauge is a CloudWatch gauge metric.
type gauge struct {
	g *generic.Gauge
}

// With implements gauge
func (g *gauge) With(labelValues ...string) metrics.Gauge {
	g.g = g.g.With(labelValues...).(*generic.Gauge)
	return g
}

// Set implements gauge
func (g *gauge) Set(value float64) {
	g.g.Set(value)
}

// Add implements gauge
func (g *gauge) Add(delta float64) {
	g.g.Add(delta)
}

// histogram is a CloudWatch histogram metric
type histogram struct {
	h *generic.Histogram
}

// With implements histogram
func (h *histogram) With(labelValues ...string) metrics.Histogram {
	h.h = h.h.With(labelValues...).(*generic.Histogram)
	return h
}

// Observe implements histogram
func (h *histogram) Observe(value float64) {
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
