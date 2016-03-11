// Package metrics provides an extensible framework to instrument your
// application. All metrics are safe for concurrent use. Considerable design
// influence has been taken from https://github.com/codahale/metrics and
// https://prometheus.io.
package metrics

// Counter is a monotonically-increasing, unsigned, 64-bit integer used to
// capture the number of times an event has occurred. By tracking the deltas
// between measurements of a counter over intervals of time, an aggregation
// layer can derive rates, acceleration, etc.
type Counter interface {
	Name() string
	With(Field) Counter
	Add(delta uint64)
}

// Gauge captures instantaneous measurements of something using signed, 64-bit
// floats. The value does not need to be monotonic.
type Gauge interface {
	Name() string
	With(Field) Gauge
	Set(value float64)
	Add(delta float64)
	Get() float64
}

// Histogram tracks the distribution of a stream of values (e.g. the number of
// milliseconds it takes to handle requests). Implementations may choose to
// add gauges for values at meaningful quantiles.
type Histogram interface {
	Name() string
	With(Field) Histogram
	Observe(value int64)
	Distribution() ([]Bucket, []Quantile)
}

// Field is a key/value pair associated with an observation for a specific
// metric. Fields may be ignored by implementations.
type Field struct {
	Key   string
	Value string
}

// Bucket is a range in a histogram which aggregates observations.
type Bucket struct {
	From  int64
	To    int64
	Count int64
}

// Quantile is a pair of quantile (0..100) and its observed maximum value.
type Quantile struct {
	Quantile int // 0..100
	Value    int64
}
