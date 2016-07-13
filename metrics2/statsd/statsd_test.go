package statsd

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics2/generic"
	"github.com/go-kit/kit/metrics2/teststat"
)

func TestHistogramAdapter(t *testing.T) {
	for _, testcase := range []struct {
		observeIn time.Duration
		reportIn  time.Duration
		unit      string
		input     float64
		want      int64
	}{
		{time.Second, time.Second, "s", 0.10, 0},
		{time.Second, time.Second, "s", 1.01, 1},
		{time.Second, time.Millisecond, "ms", 1.23, 1230},
		{time.Millisecond, time.Microsecond, "us", 123, 123000},
	} {
		tm := NewTiming(testcase.unit, 1.0)
		h := newHistogram(testcase.observeIn, testcase.reportIn, tm)
		h.Observe(testcase.input)
		if want, have := testcase.want, tm.Values()[0]; want != have {
			t.Errorf("Observe(%.2f %s): want %d, have %d", testcase.input, testcase.unit, want, have)
		}
	}
}

func TestCounter(t *testing.T) {
	prefix, name := "hello.", "world"
	re := regexp.MustCompile(prefix + name + `:([0-9\.]+)\|c`) // StatsD protocol
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
	prefix, name := "hello.", "world"
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
	prefix, name := "statsd.", "histogram_test"
	re := regexp.MustCompile(`^` + prefix + name + `:([0-9\.]+)\|ms$`)
	s := NewRaw(prefix, log.NewNopLogger())
	histogram := s.MustNewHistogram(name, time.Millisecond, time.Millisecond, 1.0)

	// Like DogStatsD, Statsd histograms (Timings) just emit all observations.
	// So, we collect them into a generic histogram, and run the statistics test
	// on that.
	quantiles := func() (float64, float64, float64, float64) {
		var buf bytes.Buffer
		s.WriteTo(&buf)
		fmt.Fprintf(os.Stderr, "%s\n", buf.String())
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

func TestTiming(t *testing.T) {
	t.Skip("TODO")
}
