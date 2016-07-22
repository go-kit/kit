// Package statsd implements a StatsD backend for package metrics. Metrics are
// aggregated and reported in batches, in the StatsD plaintext format. Sampling
// is not supported for counters because we aggregate counter updates and send
// in batches. Sampling is, however, supported for Timings.
//
// Batching observations and emitting every few seconds is useful even if you
// connect to your StatsD server over UDP. Emitting one network packet per
// observation can quickly overwhelm even the fastest internal network. Batching
// allows you to more linearly scale with growth.
//
// Typically you'll create a Statsd object in your main function.
//
//    s, stop := New("myprefix.", "udp", "statsd:8126", time.Second, log.NewNopLogger())
//    defer stop()
//
// Then, create the metrics that your application will track from that object.
// Pass them as dependencies to the component that needs them; don't place them
// in the global scope.
//
//    requests := s.NewCounter("requests")
//    depth := s.NewGauge("queue_depth")
//    fooLatency := s.NewTiming("foo_duration", "ms", 1.0)
//    barLatency := s.MustNewHistogram("bar_duration", time.Second, time.Millisecond, 1.0)
//
// Invoke them in your components when you have something to instrument.
//
//    requests.Add(1)
//    depth.Set(123)
//    fooLatency.Observe(16)    // 16 ms
//    barLatency.Observe(0.032) // 0.032 sec = 32 ms
//
package statsd

import (
	"fmt"
	"io"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics2"
	"github.com/go-kit/kit/metrics2/internal/push"
	"github.com/go-kit/kit/util/conn"
)

// Statsd is a buffer for metrics that will be emitted in the StatsD format.
// Create a Statsd object, use it to create metrics, and pass those metrics as
// dependencies to components that will use them.
//
// All observations are kept and buffered until WriteTo is called. To regularly
// report metrics to an io.Writer, use the WriteLoop helper method. To send to a
// remote StatsD server, use the SendLoop helper method.
type Statsd struct {
	buffer *push.Buffer
	logger log.Logger
}

// New returns a Statsd object that may be used to create metrics. Prefix is
// applied to all created metrics. bufSz controls the buffer depth of various
// internal channels, which can help to mitigate blocking in the observation
// path. Callers must ensure that regular calls to WriteTo are performed, either
// manually or with one of the helper methods.
func New(prefix string, bufSz int, logger log.Logger) *Statsd {
	return &Statsd{
		buffer: push.NewBuffer(prefix, bufSz),
		logger: logger,
	}
}

// NewCounter returns a new counter. Observations are collected in this object
// and flushed when WriteTo is invoked. If no sampling is required, pass
// sampleRate 1.0.
func (d *Statsd) NewCounter(name string, sampleRate float64) metrics.Counter {
	return d.buffer.NewCounter(name, sampleRate)
}

// NewGauge returns a new gauge. Observations are collected in this object and
// flushed when WriteTo is invoked.
func (d *Statsd) NewGauge(name string) metrics.Gauge {
	return d.buffer.NewGauge(name)
}

// NewTiming returns a histogram that assumes observations in milliseconds. See
// StatsD documentation for more detail. Observations are collected in this
// object and flushed when WriteTo is invoked. If no sampling is required, pass
// sampleRate 1.0.
func (d *Statsd) NewTiming(name string, sampleRate float64) metrics.Histogram {
	return d.buffer.NewHistogram(name, sampleRate)
}

// WriteLoop is a helper method that invokes WriteTo to the passed writer every
// time the passed channel fires. This method blocks until the channel is
// closed, so clients probably want to run it in its own goroutine.
func (d *Statsd) WriteLoop(c <-chan time.Time, w io.Writer) {
	for range c {
		if _, err := d.WriteTo(w); err != nil {
			d.logger.Log("during", "WriteTo", "err", err)
		}
	}
}

// SendLoop is a helper method that wraps WriteLoop, passing a managed
// connection to the network and address. Like WriteLoop, this method blocks
// until the channel is closed, so clients probably want to start it in its own
// goroutine.
func (d *Statsd) SendLoop(c <-chan time.Time, network, address string) {
	d.WriteLoop(c, conn.NewDefaultManager(network, address, d.logger))
}

// WriteTo flushes the buffered content of the metrics to the writer, in
// StatsD format. WriteTo abides best-effort semantics, so observations are
// lost if there is a problem with the write. Clients should be sure to call
// WriteTo regularly, ideally through the WriteLoop or SendLoop helper methods.
func (d *Statsd) WriteTo(w io.Writer) (int64, error) {
	// TODO(pb): As an optimization, we can aggregate in-memory for metrics with
	// the same name and label values, and write a single aggregate line.
	adds, sets, obvs := d.buffer.Get()
	var count int64
	for _, add := range adds {
		n, err := fmt.Fprintf(w, "%s:%f|c%s\n", add.Name, add.Delta, sampling(add.SampleRate))
		if err != nil {
			return count, err
		}
		count += int64(n)
	}
	for _, set := range sets {
		n, err := fmt.Fprintf(w, "%s:%f|g\n", set.Name, set.Value)
		if err != nil {
			return count, err
		}
		count += int64(n)
	}
	for _, obv := range obvs {
		n, err := fmt.Fprintf(w, "%s:%f|ms%s\n", obv.Name, obv.Value, sampling(obv.SampleRate))
		if err != nil {
			return count, err
		}
		count += int64(n)
	}
	return count, nil
}

func sampling(r float64) string {
	var sv string
	if r < 1.0 {
		sv = fmt.Sprintf("|@%f", r)
	}
	return sv
}
