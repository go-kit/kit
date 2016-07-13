package expvar

import (
	"strconv"
	"testing"

	"github.com/go-kit/kit/metrics2/teststat"
)

func TestCounter(t *testing.T) {
	counter := NewCounter("expvar_counter")
	value := func() float64 { f, _ := strconv.ParseFloat(counter.f.String(), 64); return f }
	if err := teststat.TestCounter(counter, value); err != nil {
		t.Fatal(err)
	}
}

func TestGauge(t *testing.T) {
	gauge := NewGauge("expvar_gauge")
	value := func() float64 { f, _ := strconv.ParseFloat(gauge.f.String(), 64); return f }
	if err := teststat.TestGauge(gauge, value); err != nil {
		t.Fatal(err)
	}
}

func TestHistogram(t *testing.T) {
	histogram := NewHistogram("expvar_histogram", 50)
	quantiles := func() (float64, float64, float64, float64) {
		p50, _ := strconv.ParseFloat(histogram.p50.String(), 64)
		p90, _ := strconv.ParseFloat(histogram.p90.String(), 64)
		p95, _ := strconv.ParseFloat(histogram.p95.String(), 64)
		p99, _ := strconv.ParseFloat(histogram.p99.String(), 64)
		return p50, p90, p95, p99
	}
	if err := teststat.TestHistogram(histogram, quantiles, 0.01); err != nil {
		t.Fatal(err)
	}
}

func TestWith(t *testing.T) {
	t.Skip("TODO")
}
