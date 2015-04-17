package zipkin

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/apache/thrift/lib/go/thrift"

	"github.com/peterbourgon/gokit/tracing/zipkin/thrift/gen-go/scribe"
)

// Collector represents a Zipkin trace collector, which is probably a set of
// remote endpoints.
type Collector interface {
	Collect(*Span) error
}

// ScribeCollector implements Collector by forwarding spans to a Scribe
// service in batches.
type ScribeCollector struct {
	spanc chan spanTuple
}

// NewScribeCollector returns a new Scribe-backed Collector, ready for use.
func NewScribeCollector(addr string, timeout time.Duration, batchSize int, batchTime time.Duration) (Collector, error) {
	factory := func() (scribe.Scribe, error) {
		return newScribeClient(addr, timeout)
	}
	client, err := factory()
	if err != nil {
		return nil, err
	}
	c := &ScribeCollector{
		spanc: make(chan spanTuple),
	}
	go c.loop(client, factory, batchSize, batchTime)
	return c, nil
}

// Collect implements Collector.
func (c *ScribeCollector) Collect(s *Span) error {
	e := make(chan error)
	c.spanc <- spanTuple{s, e}
	return <-e
}

func (c *ScribeCollector) loop(client scribe.Scribe, factory func() (scribe.Scribe, error), batchSize int, batchTime time.Duration) {
	batch := make([]*scribe.LogEntry, 0, batchSize)
	nextPublish := time.Now().Add(batchTime)
	publish := func() error {
		if client == nil {
			var err error
			if client, err = factory(); err != nil {
				return fmt.Errorf("during reconnect: %v", err)
			}
		}
		if rc, err := client.Log(batch); err != nil {
			client = nil
			return fmt.Errorf("during Log: %v", err)
		} else if rc != scribe.ResultCode_OK {
			// probably transient error; don't reset client
			return fmt.Errorf("remote returned %s", rc)
		}
		batch = batch[:0]
		return nil
	}

	for t := range c.spanc {
		message, err := encode(t.Span)
		if err != nil {
			t.errc <- err
			continue
		}

		batch = append(batch, &scribe.LogEntry{
			Category: "zipkin", // TODO parmeterize?
			Message:  message,
		})

		if len(batch) >= batchSize || time.Now().After(nextPublish) {
			t.errc <- publish()
			nextPublish = time.Now().Add(batchTime)
			continue
		}

		t.errc <- nil
	}
}

type spanTuple struct {
	*Span
	errc chan error
}

func newScribeClient(addr string, timeout time.Duration) (scribe.Scribe, error) {
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

func encode(s *Span) (string, error) {
	t := thrift.NewTMemoryBuffer()
	p := thrift.NewTBinaryProtocolTransport(t)
	if err := s.Encode().Write(p); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(t.Buffer.Bytes()), nil
}

// NopCollector implements Collector but performs no work.
type NopCollector struct{}

// Collect implements Collector.
func (NopCollector) Collect(*Span) error { return nil }

// MultiCollector implements Collector by sending spans to all collectors.
type MultiCollector []Collector

// Collect implements Collector.
func (c MultiCollector) Collect(s *Span) error {
	errs := []string{}
	for _, collector := range c {
		if err := collector.Collect(s); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}
