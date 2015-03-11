package statsd

// In package metrics so we can stub tick.

import (
	"bytes"
	"runtime"
	"testing"
	"time"
)

func TestCounter(t *testing.T) {
	ch := make(chan time.Time)
	tick = func(time.Duration) <-chan time.Time { return ch }
	defer func() { tick = time.Tick }()

	buf := &bytes.Buffer{}
	c := NewCounter(buf, "test_statsd_counter", time.Second)

	c.Add(1)
	c.Add(2)
	ch <- time.Now()

	for i := 0; i < 10 && buf.Len() == 0; i++ {
		time.Sleep(time.Millisecond)
	}

	if want, have := "test_statsd_counter:1|c\ntest_statsd_counter:2|c\n", buf.String(); want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}

func TestGauge(t *testing.T) {
	ch := make(chan time.Time)
	tick = func(time.Duration) <-chan time.Time { return ch }
	defer func() { tick = time.Tick }()

	buf := &bytes.Buffer{}
	g := NewGauge(buf, "test_statsd_gauge", time.Second)

	g.Add(1)          // send command
	runtime.Gosched() // yield to buffer write
	ch <- time.Now()  // signal flush
	runtime.Gosched() // yield to flush
	if want, have := "test_statsd_gauge:+1.000000|g\n", buf.String(); want != have {
		t.Errorf("want %q, have %q", want, have)
	}

	buf.Reset()

	g.Add(-2)
	runtime.Gosched()
	ch <- time.Now()
	runtime.Gosched()
	if want, have := "test_statsd_gauge:-2.000000|g\n", buf.String(); want != have {
		t.Errorf("want %q, have %q", want, have)
	}

	buf.Reset()

	g.Set(3)
	runtime.Gosched()
	ch <- time.Now()
	runtime.Gosched()
	if want, have := "test_statsd_gauge:3.000000|g\n", buf.String(); want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}

func TestHistogram(t *testing.T) {
	ch := make(chan time.Time)
	tick = func(time.Duration) <-chan time.Time { return ch }
	defer func() { tick = time.Tick }()

	buf := &bytes.Buffer{}
	h := NewHistogram(buf, "test_statsd_histogram", time.Second)

	h.Observe(123)

	runtime.Gosched()
	ch <- time.Now()
	runtime.Gosched()
	if want, have := "test_statsd_histogram:123|ms\n", buf.String(); want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}
