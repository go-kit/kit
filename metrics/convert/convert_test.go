package convert

// This is package generic_test in order to get around an import cycle: this
// package imports teststat to do its testing, but package teststat imports
// generic to use its Histogram in the Quantiles helper function.

import (
	"testing"

	"github.com/go-kit/kit/metrics/generic"
	"github.com/go-kit/kit/metrics/teststat"
)

func TestCounterHistogramConversion(t *testing.T) {
	name := "my_counter"
	c := generic.NewCounter(name)
	h := NewCounterAsHistogram(c)
	top := NewHistogramAsCounter(h).With("label", "counter").(histogramCounter)
	mid := top.h.(counterHistogram)
	low := mid.c.(*generic.Counter)
	if want, have := name, low.Name; want != have {
		t.Errorf("Name: want %q, have %q", want, have)
	}
	value := func() float64 { return low.Value() }
	if err := teststat.TestCounter(top, value); err != nil {
		t.Fatal(err)
	}
}

func TestCounterGaugeConversion(t *testing.T) {
	name := "my_counter"
	c := generic.NewCounter(name)
	g := NewCounterAsGauge(c)
	top := NewGaugeAsCounter(g).With("label", "counter").(gaugeCounter)
	mid := top.g.(counterGauge)
	low := mid.c.(*generic.Counter)
	if want, have := name, low.Name; want != have {
		t.Errorf("Name: want %q, have %q", want, have)
	}
	value := func() float64 { return low.Value() }
	if err := teststat.TestCounter(top, value); err != nil {
		t.Fatal(err)
	}
}
