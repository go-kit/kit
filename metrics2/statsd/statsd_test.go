package statsd

import (
	"testing"
	"time"
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
