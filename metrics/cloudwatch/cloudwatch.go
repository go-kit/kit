package cloudwatch

import (
	"sync"

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
	//gauges     map[string]*Gauge
	//histograms map[string]*Histogram
	logger log.Logger
}

func New(prefix string, logger log.Logger, svc *cloudwatch.CloudWatch) *CloudWatch {
	return &CloudWatch{
		prefix:   prefix,
		svc:      svc,
		counters: map[string]*Counter{},
		logger:   logger,
	}
}

func (cw *CloudWatch) NewCounter(name string) *Counter {
	c := NewCounter(cw.prefix, name)
	cw.mtx.Lock()
	cw.counters[cw.prefix+name] = c
	cw.mtx.Unlock()
	return c
}

// Counter is a Graphite counter metric.
type Counter struct {
	namespace string
	c         *generic.Counter
}

// NewCounter returns a new usable counter metric.
func NewCounter(namespace, name string) *Counter {
	return &Counter{
		namespace: namespace,
		c:         generic.NewCounter(name),
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
