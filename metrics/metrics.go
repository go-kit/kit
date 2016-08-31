package metrics

import "time"

// Counter describes a metric that accumulates values monotonically.
// An example of a counter is the number of received HTTP requests.
type Counter interface {
	With(labelValues ...string) Counter
	Add(delta float64)
}

// Gauge describes a metric that takes specific values over time.
// An example of a gauge is the current depth of a job queue.
type Gauge interface {
	With(labelValues ...string) Gauge
	Set(value float64)
}

// Histogram describes a metric that takes repeated observations of the same
// kind of thing, and produces a statistical summary of those observations,
// typically expressed as quantiles or buckets. An example of a histogram is
// HTTP request latencies.
type Histogram interface {
	With(labelValues ...string) Histogram
	Observe(value float64)
  // Start a timer used to record a duration in seconds.
  StartTimer() HistogramTimer
}


// HistogramTimer is used to implement StartTimer.
type HistogramTimer struct {
  h Histogram
  start time.Time
}

// Start a timer for the given histogram.
func NewHistogramTimer(h Histogram) HistogramTimer {
  return HistogramTimer{h: h, start: time.Now()}
}

// Stop the timer and observe the duration in seconds.
func (ht *HistogramTimer) ObserveDuration() {
  duration := time.Since(ht.start).Seconds()
  if duration < 0 {
    // Time has gone backwards.
    duration = 0
  }
  ht.h.Observe(duration)
}


