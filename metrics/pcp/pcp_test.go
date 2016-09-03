package pcp

import (
	"testing"

	"github.com/go-kit/kit/metrics/teststat"
)

func TestCounter(t *testing.T) {
	counter := NewCounter("speed_counter").With("label values", "not supported").(*Counter)
	value := func() float64 { f := counter.c.Val(); return float64(f) }
	if err := teststat.TestCounter(counter, value); err != nil {
		t.Fatal(err)
	}
}

func TestGauge(t *testing.T) {
	gauge := NewGauge("speed_gauge").With("label values", "not supported").(*Gauge)
	value := func() float64 { f := gauge.g.Val(); return f }
	if err := teststat.TestGauge(gauge, value); err != nil {
		t.Fatal(err)
	}
}

func TestHistogram(t *testing.T) {
	histogram := NewHistogram("speed_histogram").With("label values", "not supported").(*Histogram)
	quantiles := func() (float64, float64, float64, float64) {
		p50 := float64(histogram.Percentile(50))
		p90 := float64(histogram.Percentile(90))
		p95 := float64(histogram.Percentile(95))
		p99 := float64(histogram.Percentile(99))
		return p50, p90, p95, p99
	}
	if err := teststat.TestHistogram(histogram, quantiles, 0.01); err != nil {
		t.Fatal(err)
	}
}
