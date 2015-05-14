package expvar_test

import (
	stdexpvar "expvar"
	"fmt"
	"testing"

	"github.com/go-kit/kit/metrics/expvar"
	"github.com/go-kit/kit/metrics/teststat"
)

func TestHistogramQuantiles(t *testing.T) {
	metricName := "test_histogram"
	quantiles := []int{50, 90, 95, 99}
	h := expvar.NewHistogram(metricName, 0, 100, 3, quantiles...)

	const seed, mean, stdev int64 = 424242, 50, 10
	teststat.PopulateNormalHistogram(t, h, seed, mean, stdev)
	teststat.AssertExpvarNormalHistogram(t, metricName, mean, stdev, quantiles)
}

func TestCallbackGauge(t *testing.T) {
	value := 42.43
	metricName := "foo"
	expvar.PublishCallbackGauge(metricName, func() float64 { return value })
	if want, have := fmt.Sprint(value), stdexpvar.Get(metricName).String(); want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}
