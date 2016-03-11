package dogstatsd

import (
	"bytes"
	"fmt"
	"github.com/go-kit/kit/metrics"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestCounter(t *testing.T) {
	buf := &syncbuf{buf: &bytes.Buffer{}}
	reportc := make(chan time.Time)
	tags := []metrics.Field{}
	c := NewCounterTick(buf, "test_statsd_counter", reportc, tags)

	c.Add(1)
	c.With(metrics.Field{"foo", "bar"}).Add(2)
	c.With(metrics.Field{"foo", "bar"}).With(metrics.Field{"abc", "123"}).Add(2)
	c.Add(3)

	want, have := "test_statsd_counter:1|c\ntest_statsd_counter:2|c|#foo:bar\ntest_statsd_counter:2|c|#foo:bar,abc:123\ntest_statsd_counter:3|c\n", ""
	by(t, 100*time.Millisecond, func() bool {
		have = buf.String()
		return want == have
	}, func() {
		reportc <- time.Now()
	}, fmt.Sprintf("want %q, have %q", want, have))
}

func TestGauge(t *testing.T) {
	buf := &syncbuf{buf: &bytes.Buffer{}}
	reportc := make(chan time.Time)
	tags := []metrics.Field{}
	g := NewGaugeTick(buf, "test_statsd_gauge", reportc, tags)

	delta := 1.0
	g.Add(delta)

	want, have := fmt.Sprintf("test_statsd_gauge:+%f|g\n", delta), ""
	by(t, 100*time.Millisecond, func() bool {
		have = buf.String()
		return want == have
	}, func() {
		reportc <- time.Now()
	}, fmt.Sprintf("want %q, have %q", want, have))

	buf.Reset()
	delta = -2.0
	g.With(metrics.Field{"foo", "bar"}).Add(delta)

	want, have = fmt.Sprintf("test_statsd_gauge:%f|g|#foo:bar\n", delta), ""
	by(t, 100*time.Millisecond, func() bool {
		have = buf.String()
		return want == have
	}, func() {
		reportc <- time.Now()
	}, fmt.Sprintf("want %q, have %q", want, have))

	buf.Reset()
	value := 3.0
	g.With(metrics.Field{"foo", "bar"}).With(metrics.Field{"abc", "123"}).Set(value)

	want, have = fmt.Sprintf("test_statsd_gauge:%f|g|#foo:bar,abc:123\n", value), ""
	by(t, 100*time.Millisecond, func() bool {
		have = buf.String()
		return want == have
	}, func() {
		reportc <- time.Now()
	}, fmt.Sprintf("want %q, have %q", want, have))
}

func TestCallbackGauge(t *testing.T) {
	buf := &syncbuf{buf: &bytes.Buffer{}}
	reportc, scrapec := make(chan time.Time), make(chan time.Time)
	value := 55.55
	cb := func() float64 { return value }
	NewCallbackGaugeTick(buf, "test_statsd_callback_gauge", reportc, scrapec, cb)

	scrapec <- time.Now()
	reportc <- time.Now()

	// Travis is annoying
	by(t, time.Second, func() bool {
		return buf.String() != ""
	}, func() {
		reportc <- time.Now()
	}, "buffer never got write+flush")

	want, have := fmt.Sprintf("test_statsd_callback_gauge:%f|g\n", value), ""
	by(t, 100*time.Millisecond, func() bool {
		have = buf.String()
		return strings.HasPrefix(have, want) // HasPrefix because we might get multiple writes
	}, func() {
		reportc <- time.Now()
	}, fmt.Sprintf("want %q, have %q", want, have))
}

func TestHistogram(t *testing.T) {
	buf := &syncbuf{buf: &bytes.Buffer{}}
	reportc := make(chan time.Time)
	tags := []metrics.Field{}
	h := NewHistogramTick(buf, "test_statsd_histogram", reportc, tags)

	h.Observe(123)
	h.With(metrics.Field{"foo", "bar"}).Observe(456)

	want, have := "test_statsd_histogram:123|ms\ntest_statsd_histogram:456|ms|#foo:bar\n", ""
	by(t, 100*time.Millisecond, func() bool {
		have = buf.String()
		return want == have
	}, func() {
		reportc <- time.Now()
	}, fmt.Sprintf("want %q, have %q", want, have))
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

type syncbuf struct {
	mtx sync.Mutex
	buf *bytes.Buffer
}

func (s *syncbuf) Write(p []byte) (int, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	return s.buf.Write(p)
}

func (s *syncbuf) String() string {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	return s.buf.String()
}

func (s *syncbuf) Reset() {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.buf.Reset()
}
