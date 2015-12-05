package zipkin_test

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/apache/thrift/lib/go/thrift"

	"github.com/go-kit/kit/tracing/zipkin"
	"github.com/go-kit/kit/tracing/zipkin/_thrift/gen-go/scribe"
	"github.com/go-kit/kit/tracing/zipkin/_thrift/gen-go/zipkincore"
)

func TestScribeCollector(t *testing.T) {
	server := newScribeServer(t)

	timeout := time.Second
	batchInterval := time.Millisecond
	c, err := zipkin.NewScribeCollector(server.addr(), timeout, zipkin.ScribeBatchSize(0), zipkin.ScribeBatchInterval(batchInterval))
	if err != nil {
		t.Fatal(err)
	}

	var (
		serviceName  = "service"
		methodName   = "method"
		traceID      = int64(123)
		spanID       = int64(456)
		parentSpanID = int64(0)
		value        = "foo"
		duration     = 42 * time.Millisecond
	)

	span := zipkin.NewSpan("1.2.3.4:1234", serviceName, methodName, traceID, spanID, parentSpanID)
	span.AnnotateDuration("foo", 42*time.Millisecond)
	if err := c.Collect(span); err != nil {
		t.Errorf("error during collection: %v", err)
	}

	// Need to yield to the select loop to accept the send request, and then
	// yield again to the send operation to write to the socket. I think the
	// best way to do that is just give it some time.

	deadline := time.Now().Add(1 * time.Second)
	for {
		if time.Now().After(deadline) {
			t.Fatalf("never received a span")
		}
		if want, have := 1, len(server.spans()); want != have {
			time.Sleep(time.Millisecond)
			continue
		}
		break
	}

	gotSpan := server.spans()[0]
	if want, have := methodName, gotSpan.GetName(); want != have {
		t.Errorf("want %q, have %q", want, have)
	}
	if want, have := traceID, gotSpan.GetTraceId(); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
	if want, have := spanID, gotSpan.GetId(); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
	if want, have := parentSpanID, gotSpan.GetParentId(); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	if want, have := 1, len(gotSpan.GetAnnotations()); want != have {
		t.Fatalf("want %d, have %d", want, have)
	}

	gotAnnotation := gotSpan.GetAnnotations()[0]
	if want, have := value, gotAnnotation.GetValue(); want != have {
		t.Errorf("want %q, have %q", want, have)
	}
	if want, have := duration, time.Duration(gotAnnotation.GetDuration())*time.Microsecond; want != have {
		t.Errorf("want %s, have %s", want, have)
	}
}

type scribeServer struct {
	t         *testing.T
	transport *thrift.TServerSocket
	address   string
	server    *thrift.TSimpleServer
	handler   *scribeHandler
}

func newScribeServer(t *testing.T) *scribeServer {
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	transportFactory := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())

	var port int
	var transport *thrift.TServerSocket
	var err error
	for i := 0; i < 10; i++ {
		port = 10000 + rand.Intn(10000)
		transport, err = thrift.NewTServerSocket(fmt.Sprintf(":%d", port))
		if err != nil {
			t.Logf("port %d: %v", port, err)
			continue
		}
		break
	}
	if err != nil {
		t.Fatal(err)
	}

	handler := newScribeHandler(t)
	server := thrift.NewTSimpleServer4(
		scribe.NewScribeProcessor(handler),
		transport,
		transportFactory,
		protocolFactory,
	)

	go server.Serve()

	deadline := time.Now().Add(time.Second)
	for !canConnect(port) {
		if time.Now().After(deadline) {
			t.Fatal("server never started")
		}
		time.Sleep(time.Millisecond)
	}

	return &scribeServer{
		transport: transport,
		address:   fmt.Sprintf("127.0.0.1:%d", port),
		handler:   handler,
	}
}

func (s *scribeServer) addr() string {
	return s.address
}

func (s *scribeServer) spans() []*zipkincore.Span {
	return s.handler.spans()
}

type scribeHandler struct {
	t *testing.T
	sync.RWMutex
	entries []*scribe.LogEntry
}

func newScribeHandler(t *testing.T) *scribeHandler {
	return &scribeHandler{t: t}
}

func (h *scribeHandler) Log(messages []*scribe.LogEntry) (scribe.ResultCode, error) {
	h.Lock()
	defer h.Unlock()
	for _, m := range messages {
		h.entries = append(h.entries, m)
	}
	return scribe.ResultCode_OK, nil
}

func (h *scribeHandler) spans() []*zipkincore.Span {
	h.RLock()
	defer h.RUnlock()
	spans := []*zipkincore.Span{}
	for _, m := range h.entries {
		decoded, err := base64.StdEncoding.DecodeString(m.GetMessage())
		if err != nil {
			h.t.Error(err)
			continue
		}
		buffer := thrift.NewTMemoryBuffer()
		if _, err := buffer.Write(decoded); err != nil {
			h.t.Error(err)
			continue
		}
		transport := thrift.NewTBinaryProtocolTransport(buffer)
		zs := &zipkincore.Span{}
		if err := zs.Read(transport); err != nil {
			h.t.Error(err)
			continue
		}
		spans = append(spans, zs)
	}
	return spans
}

func canConnect(port int) bool {
	c, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return false
	}
	c.Close()
	return true
}
