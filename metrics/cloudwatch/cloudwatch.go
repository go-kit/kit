package cloudwatch

import (
	"sync"

	"time"

	"strconv"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/generic"
)

// CloudWatch ...
type CloudWatch struct {
	mtx        sync.RWMutex
	namespace  string
	svc        *cloudwatch.CloudWatch
	counters   map[string]*Counter
	gauges     map[string]*Gauge
	histograms map[string]*Histogram
	logger     log.Logger
}

// New ...
func New(namespace string, logger log.Logger, svc *cloudwatch.CloudWatch) *CloudWatch {
	return &CloudWatch{
		namespace:  namespace,
		svc:        svc,
		counters:   map[string]*Counter{},
		gauges:     map[string]*Gauge{},
		histograms: map[string]*Histogram{},
		logger:     logger,
	}
}

func (cw *CloudWatch) NewCounter(name string) *Counter {
	c := NewCounter(name)
	cw.mtx.Lock()
	cw.counters[name] = c
	cw.mtx.Unlock()
	return c
}

func (cw *CloudWatch) NewGauge(name string) *Gauge {
	g := NewGauge(name)
	cw.mtx.Lock()
	cw.gauges[name] = g
	cw.mtx.Unlock()
	return g
}

func (cw *CloudWatch) NewHistogram(name string, quantiles []float64, buckets int) *Histogram {
	h := NewHistogram(name, quantiles, buckets)
	cw.mtx.Lock()
	cw.histograms[name] = h
	cw.mtx.Unlock()
	return h
}

// WriteLoop is a helper method that invokes WriteTo to the passed writer every
// time the passed channel fires. This method blocks until the channel is
// closed, so clients probably want to run it in its own goroutine. For typical
// usage, create a time.Ticker and pass its C channel to this method.
func (cw *CloudWatch) WriteLoop(c <-chan time.Time) {
	for range c {
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
			for _, quantile := range h.quantiles {
				quantileStr := strconv.FormatFloat(quantile, 'f', 2, 64)
				datums = append(datums, &cloudwatch.MetricDatum{
					MetricName: aws.String(fmt.Sprintf("%s_%s", name, quantileStr)),
					Dimensions: makeDimensions(h.h.LabelValues()...),
					Value:      aws.Float64(h.h.Quantile(quantile)),
					Timestamp:  aws.Time(now),
				})
			}
		}

		_, err := cw.svc.PutMetricData(&cloudwatch.PutMetricDataInput{
			Namespace:  aws.String(cw.namespace),
			MetricData: datums,
		})
		if err != nil {
			cw.logger.Log("during", "WriteLoop", "err", err)
		}
	}
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

// With is a no-op.
func (c *Counter) With(labelValues ...string) metrics.Counter {
	return c.c.With(labelValues...)
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

func (g *Gauge) With(labelValues ...string) metrics.Gauge {
	return &Gauge{
		g: g.g.With(labelValues...).(*generic.Gauge),
	}
}

func (g *Gauge) Set(value float64) {
	g.g.Set(value)
}

func (g *Gauge) Add(delta float64) {
	g.g.Add(delta)
}

// Histogram is a CloudWatch histogram metric
type Histogram struct {
	quantiles []float64
	h         *generic.Histogram
}

// NewHistogram returns a new usable histogram metric
func NewHistogram(name string, quantiles []float64, buckets int) *Histogram {
	return &Histogram{
		quantiles: quantiles,
		h:         generic.NewHistogram(name, buckets),
	}
}

func (h *Histogram) With(labelValues ...string) metrics.Histogram {
	return &Histogram{
		h: h.h.With(labelValues...).(*generic.Histogram),
	}
}

func (h *Histogram) Observe(value float64) {
	h.h.Observe(value)
}

func makeDimensions(labelValues ...string) []*cloudwatch.Dimension {
	dimensions := make([]*cloudwatch.Dimension, len(labelValues)/2)
	for i, j := 0, 0; i < len(labelValues); i, j := i+2, j+1 {
		dimensions[j] = &cloudwatch.Dimension{
			Name:  aws.String(labelValues[i]),
			Value: aws.String(labelValues[i+1]),
		}
	}
	return dimensions
}
