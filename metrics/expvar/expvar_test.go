package expvar_test

import (
	"testing"

	"github.com/peterbourgon/gokit/metrics/expvar"
	"github.com/peterbourgon/gokit/metrics/teststat"
)

func TestHistogramQuantiles(t *testing.T) {
	metricName := "test_histogram"
	quantiles := []int{50, 90, 95, 99}
	h := expvar.NewHistogram(metricName, 0, 100, 3, quantiles...)

	const seed, mean, stdev int64 = 424242, 50, 10
	teststat.PopulateNormalHistogram(t, h, seed, mean, stdev)
	teststat.AssertExpvarNormalHistogram(t, metricName, mean, stdev, quantiles)
}
