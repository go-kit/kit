package dogstatsd

import (
	"errors"
	"testing"

	"github.com/go-kit/kit/metrics/internal/lv"
	"github.com/go-kit/kit/metrics/teststat"
	"github.com/go-kit/log"
)

func TestCounter(t *testing.T) {
	prefix, name := "abc.", "def"
	label, value := "label", "value"
	regex := `^` + prefix + name + `:([0-9\.]+)\|c\|#` + label + `:` + value + `$`
	d := New(prefix, log.NewNopLogger())
	counter := d.NewCounter(name, 1.0).With(label, value)
	valuef := teststat.SumLines(d, regex)
	if err := teststat.TestCounter(counter, valuef); err != nil {
		t.Fatal(err)
	}
}

func TestCounterSampled(t *testing.T) {
	// This will involve multiplying the observed sum by the inverse of the
	// sample rate and checking against the expected value within some
	// tolerance.
	t.Skip("TODO")
}

func TestGauge(t *testing.T) {
	prefix, name := "ghi.", "jkl"
	label, value := "xyz", "abc"
	regex := `^` + prefix + name + `:([0-9\.]+)\|g\|#hostname:foohost,` + label + `:` + value + `$`
	d := New(prefix, log.NewNopLogger(), "hostname", "foohost")
	gauge := d.NewGauge(name).With(label, value)
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
	regex := `^` + prefix + name + `:([0-9\.]+)\|h\|#` + label + `:` + value + `$`
	d := New(prefix, log.NewNopLogger())
	histogram := d.NewHistogram(name, 1.0).With(label, value)
	quantiles := teststat.Quantiles(d, regex, 50) // no |@0.X
	if err := teststat.TestHistogram(histogram, quantiles, 0.01); err != nil {
		t.Fatal(err)
	}
}

func TestHistogramSampled(t *testing.T) {
	prefix, name := "dogstatsd.", "sampled_histogram_test"
	label, value := "foo", "bar"
	regex := `^` + prefix + name + `:([0-9\.]+)\|h\|@0\.01[0]*\|#` + label + `:` + value + `$`
	d := New(prefix, log.NewNopLogger())
	histogram := d.NewHistogram(name, 0.01).With(label, value)
	quantiles := teststat.Quantiles(d, regex, 50)
	if err := teststat.TestHistogram(histogram, quantiles, 0.02); err != nil {
		t.Fatal(err)
	}
}

func TestTiming(t *testing.T) {
	prefix, name := "dogstatsd.", "timing_test"
	label, value := "wiggle", "bottom"
	regex := `^` + prefix + name + `:([0-9\.]+)\|ms\|#` + label + `:` + value + `$`
	d := New(prefix, log.NewNopLogger())
	histogram := d.NewTiming(name, 1.0).With(label, value)
	quantiles := teststat.Quantiles(d, regex, 50) // no |@0.X
	if err := teststat.TestHistogram(histogram, quantiles, 0.01); err != nil {
		t.Fatal(err)
	}
}

func TestTimingSampled(t *testing.T) {
	prefix, name := "dogstatsd.", "sampled_timing_test"
	label, value := "internal", "external"
	regex := `^` + prefix + name + `:([0-9\.]+)\|ms\|@0.03[0]*\|#` + label + `:` + value + `$`
	d := New(prefix, log.NewNopLogger())
	histogram := d.NewTiming(name, 0.03).With(label, value)
	quantiles := teststat.Quantiles(d, regex, 50)
	if err := teststat.TestHistogram(histogram, quantiles, 0.02); err != nil {
		t.Fatal(err)
	}
}

func TestDogstatsd_WriteToDiscardsMetricsOnError(t *testing.T) {
	walker := func(counter *int) func(string, lv.LabelValues, []float64) bool {
		return func(string, lv.LabelValues, []float64) bool {
			*counter++
			return true
		}
	}

	d := New("dogstatsd.", log.NewNopLogger())

	// Add some metrics.
	d.NewCounter("counter-1", 1.0).Add(1.0)
	d.NewTiming("timing-1", 1.0).Observe(1.0)
	d.NewHistogram("histogram-1", 1.0).Observe(1.0)

	// Count metrics buffered in the Dogstatsd object.
	var (
		countersCount   int
		timingsCount    int
		histogramsCount int
	)
	d.counters.Walk(walker(&countersCount))
	d.timings.Walk(walker(&timingsCount))
	d.histograms.Walk(walker(&histogramsCount))

	// Assert we have one of each type.
	if countersCount != 1 {
		t.Fatalf("expected counters count to be 1; got %d", countersCount)
	}
	if timingsCount != 1 {
		t.Fatalf("expected timings count to be 1; got %d", timingsCount)
	}
	if histogramsCount != 1 {
		t.Fatalf("expected histograms count to be 1; got %d", histogramsCount)
	}

	// Simulate an error while sending metrics.
	count, err := d.WriteTo(errorWriter{})
	if count != 0 {
		t.Fatalf("expected count to be 0; got %d", count)
	}
	if err == nil {
		t.Fatalf("expected error to be nil; got %v", err)
	}

	// Reset counters and count again.
	countersCount = 0
	timingsCount = 0
	histogramsCount = 0
	d.counters.Walk(walker(&countersCount))
	d.timings.Walk(walker(&timingsCount))
	d.histograms.Walk(walker(&histogramsCount))

	// Assert buffered metrics were cleared.
	if countersCount != 0 {
		t.Fatalf("expected counters count to be 0; got %d", countersCount)
	}
	if timingsCount != 0 {
		t.Fatalf("expected timings count to be 0; got %d", timingsCount)
	}
	if histogramsCount != 0 {
		t.Fatalf("expected histograms count to be 0; got %d", histogramsCount)
	}
}

type errorWriter struct{}

func (w errorWriter) Write([]byte) (int, error) {
	return 0, errors.New("boom")
}
