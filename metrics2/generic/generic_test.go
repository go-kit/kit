package generic

import (
	"testing"

	"github.com/go-kit/kit/metrics2/teststat"
)

func TestCounter(t *testing.T) {
	counter := NewCounter()
	value := func() float64 { return counter.Value() }
	if err := teststat.TestCounter(counter, value); err != nil {
		t.Fatal(err)
	}
}

func TestGauge(t *testing.T) {
	gauge := NewGauge()
	value := func() float64 { return gauge.Value() }
	if err := teststat.TestGauge(gauge, value); err != nil {
		t.Fatal(err)
	}
}

func TestHistogram(t *testing.T) {
	histogram := NewHistogram(50)
	quantiles := func() (float64, float64, float64, float64) {
		return histogram.Quantile(0.50), histogram.Quantile(0.90), histogram.Quantile(0.95), histogram.Quantile(0.99)
	}
	if err := teststat.TestHistogram(histogram, quantiles, 0.01); err != nil {
		t.Fatal(err)
	}
}

func TestWith(t *testing.T) {
	t.Skip("TODO")
}
