package metrics_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/peterbourgon/gokit/metrics"
	"github.com/peterbourgon/gokit/metrics/expvar"
)

func TestTimeHistogram(t *testing.T) {
	const metricName string = "test_time_histogram"
	quantiles := []int{50, 90, 99}
	h0 := expvar.NewHistogram(metricName, 0, 200, 3, quantiles...)
	h := metrics.NewTimeHistogram(time.Millisecond, h0)
	const seed, mean, stdev int64 = 321, 100, 20

	for i := 0; i < 4321; i++ {
		sample := time.Duration(rand.NormFloat64()*float64(stdev)+float64(mean)) * time.Millisecond
		h.Observe(sample)
	}

	assertExpvarNormalHistogram(t, metricName, mean, stdev, quantiles)
}
