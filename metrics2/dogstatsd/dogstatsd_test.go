package dogstatsd

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics2/generic"
	"github.com/go-kit/kit/metrics2/teststat"
)

func TestCounter(t *testing.T) {
	prefix, name := "abc.", "def"
	re := regexp.MustCompile(prefix + name + `:([0-9\.]+)\|c`) // DogStatsD protocol
	d := NewRaw(prefix, log.NewNopLogger())

	counter := d.NewCounter(name)
	value := func() float64 {
		var buf bytes.Buffer
		d.WriteTo(&buf)
		match := re.FindStringSubmatch(buf.String())
		f, _ := strconv.ParseFloat(match[1], 64)
		return f
	}

	if err := teststat.TestCounter(counter, value); err != nil {
		t.Fatal(err)
	}
}

func TestGauge(t *testing.T) {
	prefix, name := "ghi.", "jkl"
	re := regexp.MustCompile(prefix + name + `:([0-9\.]+)\|g`)
	d := NewRaw(prefix, log.NewNopLogger())

	gauge := d.NewGauge(name)
	value := func() float64 {
		var buf bytes.Buffer
		d.WriteTo(&buf)
		match := re.FindStringSubmatch(buf.String())
		f, _ := strconv.ParseFloat(match[1], 64)
		return f
	}

	if err := teststat.TestGauge(gauge, value); err != nil {
		t.Fatal(err)
	}
}

func TestHistogram(t *testing.T) {
	prefix, name := "dogstatsd.", "histogram_test"
	re := regexp.MustCompile(`^` + prefix + name + `:([0-9\.]+)\|`)
	d := NewRaw(prefix, log.NewNopLogger())
	histogram := d.NewHistogram(name, 1.0)

	// DogStatsD histograms just emit all observations. So, we collect them into
	// a generic histogram, and run the statistics test on that.
	quantiles := func() (float64, float64, float64, float64) {
		var buf bytes.Buffer
		d.WriteTo(&buf)
		h := generic.NewHistogram(50)
		for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
			match := re.FindStringSubmatch(line)
			f, _ := strconv.ParseFloat(match[1], 64)
			h.Observe(f)
		}
		return h.Quantile(0.50), h.Quantile(0.90), h.Quantile(0.95), h.Quantile(0.99)
	}

	if err := teststat.TestHistogram(histogram, quantiles, 0.01); err != nil {
		t.Fatal(err)
	}
}

func TestHistogramSampled(t *testing.T) {
	t.Skip("TODO(pb)")
}
