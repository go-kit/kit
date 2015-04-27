package statsd

// In package metrics so we can stub tick.

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
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

	delta := 1.0
	g.Add(delta)      // send command
	runtime.Gosched() // yield to buffer write
	ch <- time.Now()  // signal flush
	runtime.Gosched() // yield to flush
	if want, have := fmt.Sprintf("test_statsd_gauge:+%f|g\n", delta), buf.String(); want != have {
		t.Errorf("want %q, have %q", want, have)
	}

	buf.Reset()

	delta = -2.0
	g.Add(delta)
	runtime.Gosched()
	ch <- time.Now()
	runtime.Gosched()
	if want, have := fmt.Sprintf("test_statsd_gauge:%f|g\n", delta), buf.String(); want != have {
		t.Errorf("want %q, have %q", want, have)
	}

	buf.Reset()

	value := 3.0
	g.Set(value)
	runtime.Gosched()
	ch <- time.Now()
	runtime.Gosched()
	if want, have := fmt.Sprintf("test_statsd_gauge:%f|g\n", value), buf.String(); want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}

func TestCallbackGauge(t *testing.T) {
	ch := make(chan time.Time)
	tick = func(time.Duration) <-chan time.Time { return ch }
	defer func() { tick = time.Tick }()

	buf := &bytes.Buffer{}
	value := 55.55
	cb := func() float64 { return value }
	NewCallbackGauge(buf, "test_statsd_callback_gauge", time.Second, time.Nanosecond, cb)

	ch <- time.Now()  // signal emitter
	runtime.Gosched() // yield to emitter
	ch <- time.Now()  // signal flush
	runtime.Gosched() // yield to flush

	// Travis is annoying
	check := func() bool { return buf.String() != "" }
	execute := func() { ch <- time.Now(); runtime.Gosched(); time.Sleep(5 * time.Millisecond) }
	by(t, time.Second, check, execute, "buffer never got write+flush")

	if want, have := fmt.Sprintf("test_statsd_callback_gauge:%f|g\n", value), buf.String(); !strings.HasPrefix(have, want) {
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

func by(t *testing.T, d time.Duration, check func() bool, execute func(), msg string) {
	deadline := time.Now().Add(d)
	for !check() {
		if time.Now().After(deadline) {
			t.Fatal(msg)
		}
		execute()
	}
}
