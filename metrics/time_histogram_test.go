package metrics_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/expvar"
)

func TestTimeHistogram(t *testing.T) {
	var (
		metricName = "test_time_histogram"
		minValue   = int64(0)
		maxValue   = int64(200)
		sigfigs    = 3
		quantiles  = []int{50, 90, 99}
		h          = expvar.NewHistogram(metricName, minValue, maxValue, sigfigs, quantiles...)
		th         = metrics.NewTimeHistogram(time.Millisecond, h).With(metrics.Field{Key: "a", Value: "b"})
	)

	const seed, mean, stdev int64 = 321, 100, 20

	for i := 0; i < 4321; i++ {
		sample := time.Duration(rand.NormFloat64()*float64(stdev)+float64(mean)) * time.Millisecond
		th.Observe(sample)
	}

	assertExpvarNormalHistogram(t, metricName, mean, stdev, quantiles)
}
