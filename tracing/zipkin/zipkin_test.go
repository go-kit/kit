package zipkin_test

import (
	"sync/atomic"
	"testing"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/server"
	"github.com/go-kit/kit/tracing/zipkin"
)

func TestAnnotateServer(t *testing.T) {
	const (
		hostport    = "1.2.3.4:1234"
		serviceName = "some-service"
		methodName  = "some-method"
	)

	f := zipkin.MakeNewSpanFunc(hostport, serviceName, methodName)
	c := &countingCollector{}

	var e server.Endpoint
	e = func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil }
	e = zipkin.AnnotateServer(f, c)(e)

	if want, have := int32(0), int32(c.int32); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
	if _, err := e(context.Background(), struct{}{}); err != nil {
		t.Fatal(err)
	}
	if want, have := int32(1), int32(c.int32); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestAnnotateClient(t *testing.T) {
	t.Skip("not yet implemented")
}

type countingCollector struct{ int32 }

func (c *countingCollector) Collect(*zipkin.Span) error { atomic.AddInt32(&(c.int32), 1); return nil }
