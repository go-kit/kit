package expvar_test

import (
	stdexpvar "expvar"
	"fmt"
	"testing"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/expvar"
	"github.com/go-kit/kit/metrics/teststat"
)

func TestHistogramQuantiles(t *testing.T) {
	var (
		name      = "test_histogram"
		quantiles = []int{50, 90, 95, 99}
		h         = expvar.NewHistogram(name, 0, 100, 3, quantiles...).With(metrics.Field{Key: "ignored", Value: "field"})
	)
	const seed, mean, stdev int64 = 424242, 50, 10
	teststat.PopulateNormalHistogram(t, h, seed, mean, stdev)
	teststat.AssertExpvarNormalHistogram(t, name, mean, stdev, quantiles)
}

func TestCallbackGauge(t *testing.T) {
	var (
		name  = "foo"
		value = 42.43
	)
	expvar.PublishCallbackGauge(name, func() float64 { return value })
	if want, have := fmt.Sprint(value), stdexpvar.Get(name).String(); want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}

func TestCounter(t *testing.T) {
	var (
		name  = "m"
		value = 123
	)
	expvar.NewCounter(name).With(metrics.Field{Key: "ignored", Value: "field"}).Add(uint64(value))
	if want, have := fmt.Sprint(value), stdexpvar.Get(name).String(); want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}

func TestGauge(t *testing.T) {
	var (
		name  = "xyz"
		value = 54321
		delta = 12345
		g     = expvar.NewGauge(name).With(metrics.Field{Key: "ignored", Value: "field"})
	)
	g.Set(float64(value))
	g.Add(float64(delta))
	if want, have := fmt.Sprint(value+delta), stdexpvar.Get(name).String(); want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}

func TestInvalidQuantile(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Errorf("expected panic, got none")
		} else {
			t.Logf("got expected panic: %v", err)
		}
	}()
	expvar.NewHistogram("foo", 0.0, 100.0, 3, 50, 90, 95, 99, 101)
}
