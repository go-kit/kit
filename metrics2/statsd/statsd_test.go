package statsd

import (
	"bytes"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
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
