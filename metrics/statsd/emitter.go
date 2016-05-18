package statsd

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/util/conn"
)

// Emitter is a struct to manage connections and orchestrate the emission of
// metrics to a Statsd process.
type Emitter struct {
	prefix  string
	keyVals chan keyVal
	mgr     *conn.Manager
	logger  log.Logger
	quitc   chan chan struct{}
}

type keyVal struct {
	key string
	val string
}

func stringToKeyVal(key string, keyVals chan keyVal) chan string {
	vals := make(chan string)
	go func() {
		for val := range vals {
			keyVals <- keyVal{key: key, val: val}
		}
	}()
	return vals
}

// NewEmitter will return an Emitter that will prefix all metrics names with the
// given prefix. Once started, it will attempt to create a connection with the
// given network and address via `net.Dial` and periodically post metrics to the
// connection in the statsd protocol.
func NewEmitter(network, address string, metricsPrefix string, flushInterval time.Duration, logger log.Logger) *Emitter {
	return NewEmitterDial(net.Dial, network, address, metricsPrefix, flushInterval, logger)
}

// NewEmitterDial is the same as NewEmitter, but allows you to specify your own
// Dialer function. This is primarily useful for tests.
func NewEmitterDial(dialer conn.Dialer, network, address string, metricsPrefix string, flushInterval time.Duration, logger log.Logger) *Emitter {
	e := &Emitter{
		prefix:  metricsPrefix,
		mgr:     conn.NewManager(dialer, network, address, time.After, logger),
		logger:  logger,
		keyVals: make(chan keyVal),
		quitc:   make(chan chan struct{}),
	}
	var b bytes.Buffer
	go e.loop(flushInterval, &b)
	return e
}

// NewCounter returns a Counter that emits observations in the statsd protocol
// via the Emitter's connection manager. Observations are buffered for the
// report interval or until the buffer exceeds a max packet size, whichever
// comes first. Fields are ignored.
func (e *Emitter) NewCounter(key string) metrics.Counter {
	return &statsdCounter{
		key: e.prefix + key,
		c:   stringToKeyVal(key, e.keyVals),
	}
}

// NewHistogram returns a Histogram that emits observations in the statsd
// protocol via the Emitter's conection manager. Observations are buffered for
// the reporting interval or until the buffer exceeds a max packet size,
// whichever comes first. Fields are ignored.
//
// NewHistogram is mapped to a statsd Timing, so observations should represent
// milliseconds. If you observe in units of nanoseconds, you can make the
// translation with a ScaledHistogram:
//
//    NewScaledHistogram(statsdHistogram, time.Millisecond)
//
// You can also enforce the constraint in a typesafe way with a millisecond
// TimeHistogram:
//
//    NewTimeHistogram(statsdHistogram, time.Millisecond)
//
// TODO: support for sampling.
func (e *Emitter) NewHistogram(key string) metrics.Histogram {
	return &statsdHistogram{
		key: e.prefix + key,
		h:   stringToKeyVal(key, e.keyVals),
	}
}

// NewGauge returns a Gauge that emits values in the statsd protocol via the
// the Emitter's connection manager. Values are buffered for the report
// interval or until the buffer exceeds a max packet size, whichever comes
// first. Fields are ignored.
//
// TODO: support for sampling
func (e *Emitter) NewGauge(key string) metrics.Gauge {
	return &statsdGauge{
		key: e.prefix + key,
		g:   stringToKeyVal(key, e.keyVals),
	}
}

func (e *Emitter) loop(d time.Duration, buf *bytes.Buffer) {
	ticker := time.NewTicker(d)
	defer ticker.Stop()
	for {
		select {
		case kv := <-e.keyVals:
			fmt.Fprintf(buf, "%s:%s\n", kv.key, kv.val)
			if buf.Len() > maxBufferSize {
				e.Flush(buf)
			}

		case <-ticker.C:
			e.Flush(buf)

		case q := <-e.quitc:
			e.Flush(buf)
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

// Flush will write the given buffer to a connection provided by the Emitter's
// connection manager.
func (e *Emitter) Flush(buf *bytes.Buffer) {
	conn := e.mgr.Take()
	if conn == nil {
		e.logger.Log("during", "flush", "err", "connection unavailable")
		return
	}

	_, err := conn.Write(buf.Bytes())
	if err != nil {
		e.logger.Log("during", "flush", "err", err)
	}
	e.mgr.Put(err)
}
