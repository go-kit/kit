package zipkin_test

import (
	"encoding/base64"
	"sync"
	"testing"
	"time"

	"github.com/peterbourgon/gokit/tracing/zipkin/thrift/gen-go/zipkincore"

	"github.com/apache/thrift/lib/go/thrift"

	"github.com/peterbourgon/gokit/tracing/zipkin"
	"github.com/peterbourgon/gokit/tracing/zipkin/thrift/gen-go/scribe"
)

func TestScribeCollector(t *testing.T) {
	s := newScribeServer(t)
	defer s.close()

	c, err := zipkin.NewScribeCollector(s.addr(), 100*time.Millisecond, 0, 10*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	var (
		name         = "span-name"
		traceID      = int64(123)
		spanID       = int64(456)
		parentSpanID = int64(0)
		value        = "foo"
		duration     = 42 * time.Millisecond
	)

	span := zipkin.NewSpan("some-host", c, name, traceID, spanID, parentSpanID)
	span.AnnotateDuration("foo", 42*time.Millisecond)
	if err := span.Submit(); err != nil {
		t.Errorf("error during submit: %v", err)
	}

	if want, have := 1, len(s.spans()); want != have {
		t.Fatalf("want %d, have %d", want, have)
	}

	gotSpan := s.spans()[0]
	if want, have := name, gotSpan.GetName(); want != have {
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
	wg        sync.WaitGroup
}

func newScribeServer(t *testing.T) *scribeServer {
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	transportFactory := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	transport, err := thrift.NewTServerSocket(":0")
	if err != nil {
		t.Fatal(err)
	}
	handler := &scribeHandler{}
	server := thrift.NewTSimpleServer4(
		scribe.NewScribeProcessor(handler),
		transport,
		transportFactory,
		protocolFactory,
	)

	s := &scribeServer{t: t}
	s.wg.Add(1)
	go func() { defer s.wg.Done(); server.Serve() }()
	tickc := time.Tick(10 * time.Millisecond)
	donec := time.After(1 * time.Second)
	for !transport.IsListening() {
		select {
		case <-tickc:
			continue
		case <-donec:
			t.Fatal("server never started listening")
		}
	}

	s.transport = transport
	s.address = transport.Addr().String()
	s.server = server
	s.handler = handler
	return s
}

func (s *scribeServer) addr() string {
	return s.address
}

func (s *scribeServer) spans() []*zipkincore.Span {
	spans := []*zipkincore.Span{}
	for _, m := range *s.handler {
		decoded, err := base64.StdEncoding.DecodeString(m.GetMessage())
		if err != nil {
			s.t.Error(err)
			continue
		}
		buffer := thrift.NewTMemoryBuffer()
		if _, err := buffer.Write(decoded); err != nil {
			s.t.Error(err)
			continue
		}
		transport := thrift.NewTBinaryProtocolTransport(buffer)
		zs := &zipkincore.Span{}
		if err := zs.Read(transport); err != nil {
			s.t.Error(err)
			continue
		}
		spans = append(spans, zs)
	}
	return spans
}

func (s *scribeServer) reset() {
	s.handler.reset()
}

func (s *scribeServer) close() {
	s.transport.Close()
	s.server.Stop()
	s.wg.Wait()
}

type scribeHandler []*scribe.LogEntry

func (h *scribeHandler) Log(messages []*scribe.LogEntry) (scribe.ResultCode, error) {
	for _, m := range messages {
		(*h) = append(*h, m)
	}
	return scribe.ResultCode_OK, nil
}

func (h *scribeHandler) reset() {
	(*h) = (*h)[:0]
}

/*
type server struct {
	t        *testing.T
	wg       sync.WaitGroup
	listener net.Listener

	mu     sync.Mutex
	buffer bytes.Buffer
}

func newServer(t *testing.T) *server {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	s := &server{
		t:        t,
		wg:       wg,
		listener: listener,
		buffer:   bytes.Buffer{},
	}
	go s.loop()
	return s
}

func (s *server) loop() {
	defer s.wg.Done()
	for {
		conn, err := s.listener.Accept()
		s.t.Logf("Accept %v %v", conn.LocalAddr(), conn.RemoteAddr())
		if err != nil {
			s.t.Log(err) // closing the listener triggers an error
			return
		}
		buf := make([]byte, 8192)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				s.t.Log(err) // also OK
				break
			}
			s.buffer.Write(buf[:n])
		}
		s.t.Logf("done")
	}
}

func (s *server) addr() string {
	return s.listener.Addr().String()
}

func (s *server) buf() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buffer.Bytes() // better not mutate!
}

func (s *server) close() {
	s.listener.Close()
	//s.wg.Wait()
}
*/
