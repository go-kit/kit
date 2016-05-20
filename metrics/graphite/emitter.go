package graphite

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/util/conn"
)

// Emitter is a struct to manage connections and orchestrate the emission of
// metrics to a Graphite system.
type Emitter struct {
	mtx        sync.Mutex
	prefix     string
	mgr        *conn.Manager
	counters   []*counter
	histograms []*windowedHistogram
	gauges     []*gauge
	logger     log.Logger
	quitc      chan chan struct{}
}

// NewEmitter will return an Emitter that will prefix all metrics names with the
// given prefix. Once started, it will attempt to create a connection with the
// given network and address via `net.Dial` and periodically post metrics to the
// connection in the Graphite plaintext protocol.
func NewEmitter(network, address string, metricsPrefix string, flushInterval time.Duration, logger log.Logger) *Emitter {
	return NewEmitterDial(net.Dial, network, address, metricsPrefix, flushInterval, logger)
}

// NewEmitterDial is the same as NewEmitter, but allows you to specify your own
// Dialer function. This is primarily useful for tests.
func NewEmitterDial(dialer conn.Dialer, network, address string, metricsPrefix string, flushInterval time.Duration, logger log.Logger) *Emitter {
	e := &Emitter{
		prefix: metricsPrefix,
		mgr:    conn.NewManager(dialer, network, address, time.After, logger),
		logger: logger,
		quitc:  make(chan chan struct{}),
	}
	go e.loop(flushInterval)
	return e
}

// NewCounter returns a Counter whose value will be periodically emitted in
// a Graphite-compatible format once the Emitter is started. Fields are ignored.
func (e *Emitter) NewCounter(name string) metrics.Counter {
	e.mtx.Lock()
	defer e.mtx.Unlock()
	c := newCounter(name)
	e.counters = append(e.counters, c)
	return c
}

// NewHistogram is taken from http://github.com/codahale/metrics. It returns a
// windowed HDR histogram which drops data older than five minutes.
//
// The histogram exposes metrics for each passed quantile as gauges. Quantiles
// should be integers in the range 1..99. The gauge names are assigned by using
// the passed name as a prefix and appending "_pNN" e.g. "_p50".
//
// The values of this histogram will be periodically emitted in a
// Graphite-compatible format once the Emitter is started. Fields are ignored.
func (e *Emitter) NewHistogram(name string, minValue, maxValue int64, sigfigs int, quantiles ...int) (metrics.Histogram, error) {
	gauges := map[int]metrics.Gauge{}
	for _, quantile := range quantiles {
		if quantile <= 0 || quantile >= 100 {
			return nil, fmt.Errorf("invalid quantile %d", quantile)
		}
		gauges[quantile] = e.gauge(fmt.Sprintf("%s_p%02d", name, quantile))
	}
	h := newWindowedHistogram(name, minValue, maxValue, sigfigs, gauges, e.logger)

	e.mtx.Lock()
	defer e.mtx.Unlock()
	e.histograms = append(e.histograms, h)
	return h, nil
}

// NewGauge returns a Gauge whose value will be periodically emitted in a
// Graphite-compatible format once the Emitter is started. Fields are ignored.
func (e *Emitter) NewGauge(name string) metrics.Gauge {
	e.mtx.Lock()
	defer e.mtx.Unlock()
	return e.gauge(name)
}

func (e *Emitter) gauge(name string) metrics.Gauge {
	g := &gauge{name, 0}
	e.gauges = append(e.gauges, g)
	return g
}

func (e *Emitter) loop(d time.Duration) {
	ticker := time.NewTicker(d)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			e.Flush()

		case q := <-e.quitc:
			e.Flush()
			close(q)
			return
		}
	}
}

// Stop will flush the current metrics and close the active connection. Calling
// stop more than once is a programmer error.
func (e *Emitter) Stop() {
	q := make(chan struct{})
	e.quitc <- q
	<-q
}

// Flush will write the current metrics to the Emitter's connection in the
// Graphite plaintext protocol.
func (e *Emitter) Flush() {
	e.mtx.Lock() // one flush at a time
	defer e.mtx.Unlock()

	conn := e.mgr.Take()
	if conn == nil {
		e.logger.Log("during", "flush", "err", "connection unavailable")
		return
	}

	err := e.flush(conn)
	if err != nil {
		e.logger.Log("during", "flush", "err", err)
	}
	e.mgr.Put(err)
}

func (e *Emitter) flush(w io.Writer) error {
	bw := bufio.NewWriter(w)

	for _, c := range e.counters {
		c.flush(bw, e.prefix)
	}

	for _, h := range e.histograms {
		h.flush(bw, e.prefix)
	}

	for _, g := range e.gauges {
		g.flush(bw, e.prefix)
	}

	return bw.Flush()
}
