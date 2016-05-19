package dogstatsd

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/util/conn"
)

func TestEmitterCounter(t *testing.T) {
	e, buf := testEmitter()

	c := e.NewCounter("test_statsd_counter")
	c.Add(1)
	c.Add(2)

	// give time for things to emit
	time.Sleep(time.Millisecond * 250)
	// force a flush and stop
	e.Stop()

	want := "prefix.test_statsd_counter:1|c\nprefix.test_statsd_counter:2|c\n"
	have := buf.String()
	if want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}

func TestEmitterGauge(t *testing.T) {
	e, buf := testEmitter()

	g := e.NewGauge("test_statsd_gauge")

	delta := 1.0
	g.Add(delta)

	// give time for things to emit
	time.Sleep(time.Millisecond * 250)
	// force a flush and stop
	e.Stop()

	want := fmt.Sprintf("prefix.test_statsd_gauge:+%f|g\n", delta)
	have := buf.String()
	if want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}

func TestEmitterHistogram(t *testing.T) {
	e, buf := testEmitter()
	h := e.NewHistogram("test_statsd_histogram")

	h.Observe(123)

	// give time for things to emit
	time.Sleep(time.Millisecond * 250)
	// force a flush and stop
	e.Stop()

	want := "prefix.test_statsd_histogram:123|ms\n"
	have := buf.String()
	if want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}

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

func testEmitter() (*Emitter, *syncbuf) {
	buf := &syncbuf{buf: &bytes.Buffer{}}
	e := &Emitter{
		prefix:  "prefix.",
		mgr:     conn.NewManager(mockDialer(buf), "", "", time.After, log.NewNopLogger()),
		logger:  log.NewNopLogger(),
		keyVals: make(chan keyVal),
		quitc:   make(chan chan struct{}),
	}
	go e.loop(time.Millisecond * 20)
	return e, buf
}

func mockDialer(buf *syncbuf) conn.Dialer {
	return func(net, addr string) (net.Conn, error) {
		return &mockConn{buf}, nil
	}
}

type mockConn struct {
	buf *syncbuf
}

func (c *mockConn) Read(b []byte) (n int, err error) {
	panic("not implemented")
}

func (c *mockConn) Write(b []byte) (n int, err error) {
	return c.buf.Write(b)
}

func (c *mockConn) Close() error {
	panic("not implemented")
}

func (c *mockConn) LocalAddr() net.Addr {
	panic("not implemented")
}

func (c *mockConn) RemoteAddr() net.Addr {
	panic("not implemented")
}

func (c *mockConn) SetDeadline(t time.Time) error {
	panic("not implemented")
}

func (c *mockConn) SetReadDeadline(t time.Time) error {
	panic("not implemented")
}

func (c *mockConn) SetWriteDeadline(t time.Time) error {
	panic("not implemented")
}
