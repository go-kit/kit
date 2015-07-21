package metrics_test

import (
	"testing"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/expvar"
)

func TestScaledHistogram(t *testing.T) {
	var (
		quantiles  = []int{50, 90, 99}
		scale      = int64(10)
		metricName = "test_scaled_histogram"
	)

	var h metrics.Histogram
	h = expvar.NewHistogram(metricName, 0, 1000, 3, quantiles...)
	h = metrics.NewScaledHistogram(h, scale)
	h = h.With(metrics.Field{Key: "a", Value: "b"})

	const seed, mean, stdev = 333, 500, 100          // input values
	populateNormalHistogram(t, h, seed, mean, stdev) // will be scaled down
	assertExpvarNormalHistogram(t, metricName, mean/scale, stdev/scale, quantiles)
}
