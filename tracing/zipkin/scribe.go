package zipkin

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/apache/thrift/lib/go/thrift"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/zipkin/_thrift/gen-go/scribe"
)

const defaultScribeCategory = "zipkin"

// defaultBatchInterval in seconds
const defaultBatchInterval = 1

// ScribeCollector implements Collector by forwarding spans to a Scribe
// service, in batches.
type ScribeCollector struct {
	client        scribe.Scribe
	factory       func() (scribe.Scribe, error)
	spanc         chan *Span
	sendc         chan struct{}
	batch         []*scribe.LogEntry
	nextSend      time.Time
	batchInterval time.Duration
	batchSize     int
	shouldSample  Sampler
	logger        log.Logger
	category      string
	quit          chan struct{}
}

// NewScribeCollector returns a new Scribe-backed Collector. addr should be a
// TCP endpoint of the form "host:port". timeout is passed to the Thrift dial
// function NewTSocketFromAddrTimeout. batchSize and batchInterval control the
// maximum size and interval of a batch of spans; as soon as either limit is
// reached, the batch is sent. The logger is used to log errors, such as batch
// send failures; users should provide an appropriate context, if desired.
func NewScribeCollector(addr string, timeout time.Duration, options ...ScribeOption) (Collector, error) {
	factory := scribeClientFactory(addr, timeout)
	client, err := factory()
	if err != nil {
		return nil, err
	}
	c := &ScribeCollector{
		client:        client,
		factory:       factory,
		spanc:         make(chan *Span),
		sendc:         make(chan struct{}),
		batch:         []*scribe.LogEntry{},
		batchInterval: defaultBatchInterval * time.Second,
		batchSize:     100,
		shouldSample:  SampleRate(1.0, rand.Int63()),
		logger:        log.NewNopLogger(),
		category:      defaultScribeCategory,
		quit:          make(chan struct{}),
	}
	for _, option := range options {
		option(c)
	}
	c.nextSend = time.Now().Add(c.batchInterval)
	go c.loop()
	return c, nil
}

// Collect implements Collector.
func (c *ScribeCollector) Collect(s *Span) error {
	c.spanc <- s
	return nil // accepted
}

// Close implements Collector.
func (c *ScribeCollector) Close() error {
	close(c.quit)
	return nil
}

func (c *ScribeCollector) loop() {
	tickc := time.Tick(c.batchInterval / 10)

	for {
		select {
		case span := <-c.spanc:
			if !c.shouldSample(span.traceID) {
				continue
			}
			c.batch = append(c.batch, &scribe.LogEntry{
				Category: c.category,
				Message:  scribeSerialize(span),
			})
			if len(c.batch) >= c.batchSize {
				go c.sendNow()
			}

		case <-tickc:
			if time.Now().After(c.nextSend) {
				go c.sendNow()
			}

		case <-c.sendc:
			c.nextSend = time.Now().Add(c.batchInterval)
			if err := c.send(c.batch); err != nil {
				c.logger.Log("err", err.Error())
			}
			c.batch = c.batch[:0]
		case <-c.quit:
			return
		}
	}
}

func (c *ScribeCollector) sendNow() {
	c.sendc <- struct{}{}
}

func (c *ScribeCollector) send(batch []*scribe.LogEntry) error {
	if c.client == nil {
		var err error
		if c.client, err = c.factory(); err != nil {
			return fmt.Errorf("during reconnect: %v", err)
		}
	}
	if rc, err := c.client.Log(c.batch); err != nil {
		c.client = nil
		return fmt.Errorf("during Log: %v", err)
	} else if rc != scribe.ResultCode_OK {
		// probably transient error; don't reset client
		return fmt.Errorf("remote returned %s", rc)
	}
	return nil
}

// ScribeOption sets a parameter for the StdlibAdapter.
type ScribeOption func(s *ScribeCollector)

// ScribeBatchSize sets the maximum batch size, after which a collect will be
// triggered. The default batch size is 100 traces.
func ScribeBatchSize(n int) ScribeOption {
	return func(s *ScribeCollector) { s.batchSize = n }
}

// ScribeBatchInterval sets the maximum duration we will buffer traces before
// emitting them to the collector. The default batch interval is 1 second.
func ScribeBatchInterval(d time.Duration) ScribeOption {
	return func(s *ScribeCollector) { s.batchInterval = d }
}

// ScribeSampleRate sets the sample rate used to determine if a trace will be
// sent to the collector. By default, the sample rate is 1.0, i.e. all traces
// are sent.
func ScribeSampleRate(sr Sampler) ScribeOption {
	return func(s *ScribeCollector) { s.shouldSample = sr }
}

// ScribeLogger sets the logger used to report errors in the collection
// process. By default, a no-op logger is used, i.e. no errors are logged
// anywhere. It's important to set this option in a production service.
func ScribeLogger(logger log.Logger) ScribeOption {
	return func(s *ScribeCollector) { s.logger = logger }
}

// ScribeCategory sets the Scribe category used to transmit the spans.
func ScribeCategory(category string) ScribeOption {
	return func(s *ScribeCollector) { s.category = category }
}

func scribeClientFactory(addr string, timeout time.Duration) func() (scribe.Scribe, error) {
	return func() (scribe.Scribe, error) {
		a, err := net.ResolveTCPAddr("tcp", addr)
		if err != nil {
			return nil, err
		}
		socket := thrift.NewTSocketFromAddrTimeout(a, timeout)
		transport := thrift.NewTFramedTransport(socket)
		if err := transport.Open(); err != nil {
			socket.Close()
			return nil, err
		}
		proto := thrift.NewTBinaryProtocolTransport(transport)
		client := scribe.NewScribeClientProtocol(transport, proto, proto)
		return client, nil
	}
}

func scribeSerialize(s *Span) string {
	t := thrift.NewTMemoryBuffer()
	p := thrift.NewTBinaryProtocolTransport(t)
	if err := s.Encode().Write(p); err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(t.Buffer.Bytes())
}
