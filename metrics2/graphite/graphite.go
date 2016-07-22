// Package graphite provides a Graphite backend for metrics. Metrics are emitted
// with each observation in the plaintext protocol. See
// http://graphite.readthedocs.io/en/latest/feeding-carbon.html#the-plaintext-protocol
// for more information.
//
// Graphite does not have a native understanding of metric parameterization, so
// label values are aggregated but not reported. Use distinct metrics for each
// unique combination of label values.
package graphite

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics2"
	"github.com/go-kit/kit/metrics2/generic"
	"github.com/go-kit/kit/metrics2/internal/push"
	"github.com/go-kit/kit/util/conn"
)

// Graphite is a buffer for metrics that will be emitted in the Graphite format.
// Create a Graphite object, use it to create metrics, and pass those metrics as
// dependencies to components that will use them.
//
// All observations are kept and buffered until WriteTo is called. To regularly
// report metrics to an io.Writer, use the WriteLoop helper method. To send to a
// remote Graphite server, use the SendLoop helper method.
//
// Histograms have their observations collected into a generic.Histogram. With
// every flush, the 50th, 90th, 95th, and 99th quantiles are computed and
// reported in metrics with the provided name concatenated with a suffix of the
// form ".pNN" e.g. ".p99".
type Graphite struct {
	mtx        sync.RWMutex
	prefix     string
	buffer     *push.Buffer
	histograms map[string]*Histogram
	logger     log.Logger
}

// New returns a Graphite object that may be used to create metrics. Prefix is
// applied to all created metrics. bufSz controls the buffer depth of various
// internal channels, which can help to mitigate blocking in the observation
// path. Callers must ensure that regular calls to WriteTo are performed, either
// manually or with one of the helper methods.
func New(prefix string, bufSz int, logger log.Logger) *Graphite {
	return &Graphite{
		prefix:     prefix,
		buffer:     push.NewBuffer(prefix, bufSz),
		histograms: map[string]*Histogram{},
		logger:     logger,
	}
}

// NewCounter returns a new counter. Observations are collected in this object
// and flushed when WriteTo is invoked.
func (g *Graphite) NewCounter(name string) metrics.Counter {
	return g.buffer.NewCounter(name, 1.0)
}

// NewGauge returns a new gauge. Observations are collected in this object and
// flushed when WriteTo is invoked.
func (g *Graphite) NewGauge(name string) metrics.Gauge {
	return g.buffer.NewGauge(name)
}

// NewHistogram returns a new histogram. Observations are collected into a
// generic.Histogram, and quantiles are reported with every WriteTo. Buckets
// sets the number of buckets in the histogram; 50 is a good default.
func (g *Graphite) NewHistogram(name string, buckets int) metrics.Histogram {
	// The push.Buffer batches and emits individual observations, but that's
	// intended for push-based systems where individual observations are taken
	// as input. Graphite, in contrast, has no native support for histograms,
	// timings, or anything that accepts individual observations. Instead, we
	// perform statistical aggregation in the client, and report something like
	// gauges at each quantile.
	//
	// Note that this only works because Graphite doesn't support label values.
	// That means the With method on the returned generic.Histogram is effectively
	// ignored, and observations
	g.mtx.Lock()
	defer g.mtx.Unlock()
	h := &Histogram{generic.NewHistogram(buckets)}
	g.histograms[g.prefix+name] = h
	return h
}

// WriteLoop is a helper method that invokes WriteTo to the passed writer every
// time the passed channel fires. This method blocks until the channel is
// closed, so clients probably want to run it in its own goroutine.
func (g *Graphite) WriteLoop(c <-chan time.Time, w io.Writer) {
	for range c {
		if _, err := g.WriteTo(w); err != nil {
			g.logger.Log("during", "WriteTo", "err", err)
		}
	}
}

// SendLoop is a helper method that wraps WriteLoop, passing a managed
// connection to the network and address. Like WriteLoop, this method blocks
// until the channel is closed, so clients probably want to start it in its own
// goroutine.
func (g *Graphite) SendLoop(c <-chan time.Time, network, address string) {
	g.WriteLoop(c, conn.NewDefaultManager(network, address, g.logger))
}

// WriteTo flushes the buffered content of the metrics to the writer, in
// DogStatsD format. WriteTo abides best-effort semantics, so observations are
// lost if there is a problem with the write. Clients should be sure to call
// WriteTo regularly, ideally through the WriteLoop or SendLoop helper methods.
func (g *Graphite) WriteTo(w io.Writer) (int64, error) {
	// TODO(pb): As an optimization, we can aggregate in-memory for metrics with
	// the same name and label values, and write a single aggregate line.
	adds, sets, _ := g.buffer.Get()
	now := time.Now().Unix()
	var count int64
	for _, add := range adds {
		n, err := fmt.Fprintf(w, "%s %f %d\n", add.Name, add.Delta, now)
		if err != nil {
			return count, err
		}
		count += int64(n)
	}
	for _, set := range sets {
		n, err := fmt.Fprintf(w, "%s %f %d\n", set.Name, set.Value, now)
		if err != nil {
			return count, err
		}
		count += int64(n)
	}

	// Histograms are different than the rest.
	g.mtx.RLock()
	defer g.mtx.RUnlock()
	for path, h := range g.histograms {
		for _, p := range []struct {
			s string
			f float64
		}{
			{"50", 0.50},
			{"90", 0.90},
			{"95", 0.95},
			{"99", 0.99},
		} {
			n, err := fmt.Fprintf(w, "%s.p%s %f %d\n", path, p.s, h.Quantile(p.f), now)
			if err != nil {
				return count, err
			}
			count += int64(n)
		}
	}
	return count, nil
}

// Histogram adapts the generic.Histogram and makes the With method a no-op.
// Graphite doesn't support label values, so this is fine from a protocol
// perspective. It's also necessary because histograms are registered in the
// Graphite object, and if With returns a new copy of the metric, observations
// are lost.
type Histogram struct {
	*generic.Histogram
}

// With is a no-op.
func (h *Histogram) With(...string) metrics.Histogram { return h }
