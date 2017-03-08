package cloudwatch

import (
	"sync"

	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/generic"
)

// CloudWatch ...
type CloudWatch struct {
	mtx      sync.RWMutex
	prefix   string
	svc      *cloudwatch.CloudWatch
	counters map[string]*Counter
	gauges   map[string]*Gauge
	logger   log.Logger
}

// New ...
func New(prefix string, logger log.Logger, svc *cloudwatch.CloudWatch) *CloudWatch {
	return &CloudWatch{
		prefix:   prefix,
		svc:      svc,
		counters: map[string]*Counter{},
		gauges:   map[string]*Gauge{},
		logger:   logger,
	}
}

func (cw *CloudWatch) NewCounter(name string) *Counter {
	c := NewCounter(name)
	cw.mtx.Lock()
	cw.counters[cw.prefix+name] = c
	cw.mtx.Unlock()
	return c
}

func (cw *CloudWatch) NewGauge(name string) *Gauge {
	c := NewGauge(name)
	cw.mtx.Lock()
	cw.counters[cw.prefix+name] = c
	cw.mtx.Unlock()
	return c
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

		for name, c := range cw.counters {
			_, err := cw.svc.PutMetricData(&cloudwatch.PutMetricDataInput{
				Namespace: aws.String(cw.prefix),
				MetricData: []*cloudwatch.MetricDatum{
					{
						MetricName: aws.String(name),
						Dimensions: makeDimensions(c.c.LabelValues()...),
						Value:      aws.Float64(c.c.Value()),
						Timestamp:  aws.Time(now),
					},
				},
			})
			if err != nil {
				cw.logger.Log("during", "WriteLoop", "err", err)
			}
		}

		for name, g := range cw.gauges {
			_, err := cw.svc.PutMetricData(&cloudwatch.PutMetricDataInput{
				Namespace: aws.String(cw.prefix),
				MetricData: []*cloudwatch.MetricDatum{
					{
						MetricName: aws.String(name),
						Dimensions: makeDimensions(g.g.LabelValues()...),
						Value:      aws.Float64(g.g.Value()),
						Timestamp:  aws.Time(now),
					},
				},
			})
			if err != nil {
				cw.logger.Log("during", "WriteLoop", "err", err)
			}
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
		g: g.g.With(labelValues...),
	}
}

func (g *Gauge) Set(value float64) {
	g.g.Set(value)
}

func (g *Gauge) Add(delta float64) {
	g.g.Add(delta)
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
