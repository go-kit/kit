package teststat

import (
	"expvar"
	"fmt"
	"math"
	"strconv"
	"testing"
)

// AssertExpvarNormalHistogram ensures the expvar Histogram referenced by
// metricName abides a normal distribution.
func AssertExpvarNormalHistogram(t *testing.T, metricName string, mean, stdev int64, quantiles []int) {
	const tolerance int = 2
	for _, quantile := range quantiles {
		want := normalValueAtQuantile(mean, stdev, quantile)
		s := expvar.Get(fmt.Sprintf("%s_p%02d", metricName, quantile)).String()
		have, err := strconv.Atoi(s)
		if err != nil {
			t.Fatal(err)
		}
		if int(math.Abs(float64(want)-float64(have))) > tolerance {
			t.Errorf("quantile %d: want %d, have %d", quantile, want, have)
		}
	}
}
