package generic

import (
	"testing"

	"github.com/go-kit/kit/metrics2/teststat"
)

func TestHistogram(t *testing.T) {
	histogram := NewHistogram(50)
	quantiles := func() (float64, float64, float64, float64) {
		return histogram.Quantile(0.50), histogram.Quantile(0.90), histogram.Quantile(0.95), histogram.Quantile(0.99)
	}
	if err := teststat.TestHistogram(histogram, quantiles, 0.01); err != nil {
		t.Fatal(err)
	}
}
