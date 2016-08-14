// Package dogstatsd provides a DogStatsD backend for package metrics. It's very
// similar to statsd, but supports arbitrary tags per-metric, which map to Go
// kit's label values. So, while label values are no-ops in StatsD, they are
// supported here. For more details, see the documentation at
// http://docs.datadoghq.com/guides/dogstatsd/.
//
// This package batches observations and emits them on some schedule to the
// remote server. This is useful even if you connect to your DogStatsD server
// over UDP. Emitting one network packet per observation can quickly overwhelm
// even the fastest internal network. Batching allows you to more linearly scale
// with growth.
package dogstatsd

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics2"
	"github.com/go-kit/kit/metrics2/internal/push"
	"github.com/go-kit/kit/util/conn"
)

// Dogstatsd is a buffer for metrics that will be emitted in the DogStatsD
// format. Create a Dogstatsd object, use it to create metrics, and pass those
// metrics as dependencies to components that will use them.
//
// All observations are kept and buffered until WriteTo is called. To regularly
// report metrics to an io.Writer, use the WriteLoop helper method. To send to a
// remote DogStatsD server, use the SendLoop helper method.
type Dogstatsd struct {
	prefix  string
	buffer  *push.Buffer
	timings *set // to differentiate from histograms
	logger  log.Logger
}

// New returns a Dogstatsd object that may be used to create metrics. Prefix is
// applied to all created metrics. bufSz controls the buffer depth of various
// internal channels, which can help to mitigate blocking in the observation
// path. Callers must ensure that regular calls to WriteTo are performed, either
// manually or with one of the helper methods.
func New(prefix string, bufSz int, logger log.Logger) *Dogstatsd {
	return &Dogstatsd{
		prefix:  prefix,
		buffer:  push.NewBuffer(prefix, bufSz),
		timings: newSet(),
		logger:  logger,
	}
}

// NewCounter returns a new counter. Observations are collected in this object
// and flushed when WriteTo is invoked. If no sampling is required, pass
// sampleRate 1.0.
func (d *Dogstatsd) NewCounter(name string, sampleRate float64) metrics.Counter {
	return d.buffer.NewCounter(name, sampleRate)
}

// NewGauge returns a new gauge. Observations are collected in this object and
// flushed when WriteTo is invoked.
func (d *Dogstatsd) NewGauge(name string) metrics.Gauge {
	return d.buffer.NewGauge(name)
}

// NewHistogram returns a new histogram. Observations are collected in this
// object and flushed when WriteTo is invoked. If no sampling is required, pass
// sampleRate 1.0.
func (d *Dogstatsd) NewHistogram(name string, sampleRate float64) metrics.Histogram {
	return d.buffer.NewHistogram(name, sampleRate)
}

// NewTiming is like NewHistogram but assumes observations in milliseconds.
// See DogStatsD documentation for more detail.
func (d *Dogstatsd) NewTiming(name string, sampleRate float64) metrics.Histogram {
	d.timings.add(d.prefix + name)
	return d.buffer.NewHistogram(name, sampleRate)
}

// WriteLoop is a helper method that invokes WriteTo to the passed writer every
// time the passed channel fires. This method blocks until the channel is
// closed, so clients probably want to run it in its own goroutine. For typical
// usage, create a time.Ticker and pass its C channel to this method.
func (d *Dogstatsd) WriteLoop(c <-chan time.Time, w io.Writer) {
	for range c {
		if _, err := d.WriteTo(w); err != nil {
			d.logger.Log("during", "WriteTo", "err", err)
		}
	}
}

// SendLoop is a helper method that wraps WriteLoop, passing a managed
// connection to the network and address. Like WriteLoop, this method blocks
// until the channel is closed, so clients probably want to start it in its own
// goroutine. For typical usage, create a time.Ticker and pass its C channel to
// this method.
func (d *Dogstatsd) SendLoop(c <-chan time.Time, network, address string) {
	d.WriteLoop(c, conn.NewDefaultManager(network, address, d.logger))
}

// WriteTo flushes the buffered content of the metrics to the writer, in
// DogStatsD format. WriteTo abides best-effort semantics, so observations are
// lost if there is a problem with the write. Clients should be sure to call
// WriteTo regularly, ideally through the WriteLoop or SendLoop helper methods.
func (d *Dogstatsd) WriteTo(w io.Writer) (int64, error) {
	// TODO(pb): As an optimization, we can aggregate in-memory for metrics with
	// the same name and label values, and write a single aggregate line.
	adds, sets, obvs := d.buffer.Get()
	var count int64
	for _, add := range adds {
		n, err := fmt.Fprintf(w, "%s:%f|c%s%s\n", add.Name, add.Delta, sampling(add.SampleRate), tagValues(add.LVs))
		if err != nil {
			return count, err
		}
		count += int64(n)
	}
	for _, set := range sets {
		n, err := fmt.Fprintf(w, "%s:%f|g%s\n", set.Name, set.Value, tagValues(set.LVs))
		if err != nil {
			return count, err
		}
		count += int64(n)
	}
	for _, obv := range obvs {
		suffix := "|h"
		if d.timings.has(obv.Name) {
			suffix = "|ms"
		}
		n, err := fmt.Fprintf(w, "%s:%f%s%s%s\n", obv.Name, obv.Value, suffix, sampling(obv.SampleRate), tagValues(obv.LVs))
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

func tagValues(labelValues []string) string {
	if len(labelValues) == 0 {
		return ""
	}
	if len(labelValues)%2 != 0 {
		panic("tagValues received a labelValues with an odd number of strings")
	}
	pairs := make([]string, 0, len(labelValues)/2)
	for i := 0; i < len(labelValues); i += 2 {
		pairs = append(pairs, labelValues[i]+":"+labelValues[i+1])
	}
	return "|#" + strings.Join(pairs, ",")
}

type set struct {
	mtx sync.RWMutex
	m   map[string]struct{}
}

func newSet() *set               { return &set{m: map[string]struct{}{}} }
func (s *set) add(a string)      { s.mtx.Lock(); defer s.mtx.Unlock(); s.m[a] = struct{}{} }
func (s *set) has(a string) bool { s.mtx.RLock(); defer s.mtx.RUnlock(); _, ok := s.m[a]; return ok }
