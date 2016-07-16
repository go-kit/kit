package dogstatsd

import (
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics2/teststat"
)

func TestCounter(t *testing.T) {
	prefix, name := "abc.", "def"
	label, value := "label", "value"
	regex := `^` + prefix + name + `:([0-9\.]+)\|c|#` + label + `:` + value + `$`
	d := New(prefix, 0, log.NewNopLogger())
	counter := d.NewCounter(name, 1.0)
	valuef := teststat.SumLines(d, regex)
	if err := teststat.TestCounter(counter, valuef); err != nil {
		t.Fatal(err)
	}
}

func TestCounterSampled(t *testing.T) {
	// THis will involve multiplying the observed sum by the inverse of the
	// sample rate and checking against the expected value within some
	// tolerance.
	t.Skip("TODO")
}

func TestGauge(t *testing.T) {
	prefix, name := "ghi.", "jkl"
	label, value := "xyz", "abc"
	regex := `^` + prefix + name + `:([0-9\.]+)\|g|#` + label + `:` + value + `$`
	d := New(prefix, 0, log.NewNopLogger())
	gauge := d.NewGauge(name)
	valuef := teststat.LastLine(d, regex)
	if err := teststat.TestGauge(gauge, valuef); err != nil {
		t.Fatal(err)
	}
}

// DogStatsD histograms just emit all observations. So, we collect them into
// a generic histogram, and run the statistics test on that.

func TestHistogram(t *testing.T) {
	prefix, name := "dogstatsd.", "histogram_test"
	label, value := "abc", "def"
	regex := `^` + prefix + name + `:([0-9\.]+)\|h|#` + label + `:` + value + `$`
	d := New(prefix, 0, log.NewNopLogger())
	histogram := d.NewHistogram(name, 1.0)
	quantiles := teststat.Quantiles(d, regex, 50) // no |@0.X
	if err := teststat.TestHistogram(histogram, quantiles, 0.01); err != nil {
		t.Fatal(err)
	}
}

func TestHistogramSampled(t *testing.T) {
	prefix, name := "dogstatsd.", "sampled_histogram_test"
	label, value := "foo", "bar"
	regex := `^` + prefix + name + `:([0-9\.]+)\|h|@0.1[0]*|#` + label + `:` + value + `$`
	d := New(prefix, 0, log.NewNopLogger())
	histogram := d.NewHistogram(name, 0.01).With(label, value)
	quantiles := teststat.Quantiles(d, regex, 50)
	if err := teststat.TestHistogram(histogram, quantiles, 0.02); err != nil {
		t.Fatal(err)
	}
}

func TestTiming(t *testing.T) {
	t.Skip("TODO")
}
