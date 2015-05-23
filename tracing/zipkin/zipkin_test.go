package zipkin_test

import (
	"net/http"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/client"
	"github.com/go-kit/kit/server"
	"github.com/go-kit/kit/tracing/zipkin"
)

func TestToContext(t *testing.T) {
	const (
		hostport           = "5.5.5.5:5555"
		serviceName        = "foo-service"
		methodName         = "foo-method"
		traceID      int64 = 12
		spanID       int64 = 34
		parentSpanID int64 = 56
	)

	r, _ := http.NewRequest("GET", "https://best.horse", nil)
	r.Header.Set("X-B3-TraceId", strconv.FormatInt(traceID, 16))
	r.Header.Set("X-B3-SpanId", strconv.FormatInt(spanID, 16))
	r.Header.Set("X-B3-ParentSpanId", strconv.FormatInt(parentSpanID, 16))

	newSpan := zipkin.MakeNewSpanFunc(hostport, serviceName, methodName)
	toContext := zipkin.ToContext(newSpan)

	ctx := toContext(context.Background(), r)
	val := ctx.Value(zipkin.SpanContextKey)
	if val == nil {
		t.Fatalf("%s returned no value", zipkin.SpanContextKey)
	}
	span, ok := val.(*zipkin.Span)
	if !ok {
		t.Fatalf("%s was not a Span object", zipkin.SpanContextKey)
	}

	for want, haveFunc := range map[int64]func() int64{
		traceID:      span.TraceID,
		spanID:       span.SpanID,
		parentSpanID: span.ParentSpanID,
	} {
		if have := haveFunc(); want != have {
			name := runtime.FuncForPC(reflect.ValueOf(haveFunc).Pointer()).Name()
			name = strings.Split(name, "·")[0]
			toks := strings.Split(name, ".")
			name = toks[len(toks)-1]
			t.Errorf("%s: want %d, have %d", name, want, have)
		}
	}
}

func TestToRequest(t *testing.T) {
	const (
		hostport           = "5.5.5.5:5555"
		serviceName        = "foo-service"
		methodName         = "foo-method"
		traceID      int64 = 20
		spanID       int64 = 40
		parentSpanID int64 = 90
	)

	newSpan := zipkin.MakeNewSpanFunc(hostport, serviceName, methodName)
	span := newSpan(traceID, spanID, parentSpanID)
	ctx := context.WithValue(context.Background(), zipkin.SpanContextKey, span)
	r, _ := http.NewRequest("GET", "https://best.horse", nil)
	ctx = zipkin.ToRequest(newSpan)(ctx, r)

	for header, wantInt := range map[string]int64{
		"X-B3-TraceId":      traceID,
		"X-B3-SpanId":       spanID,
		"X-B3-ParentSpanId": parentSpanID,
	} {
		if want, have := strconv.FormatInt(wantInt, 16), r.Header.Get(header); want != have {
			t.Errorf("%s: want %q, have %q", header, want, have)
		}
	}
}

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

	if want, have := int32(0), atomic.LoadInt32(&(c.int32)); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
	if _, err := e(context.Background(), struct{}{}); err != nil {
		t.Fatal(err)
	}
	if want, have := int32(1), atomic.LoadInt32(&(c.int32)); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestAnnotateClient(t *testing.T) {
	const (
		hostport    = "192.168.1.100:53"
		serviceName = "client-service"
		methodName  = "client-method"
	)

	f := zipkin.MakeNewSpanFunc(hostport, serviceName, methodName)
	c := &countingCollector{}

	var e client.Endpoint
	e = func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil }
	e = zipkin.AnnotateClient(f, c)(e)

	if want, have := int32(0), atomic.LoadInt32(&(c.int32)); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
	if _, err := e(context.Background(), struct{}{}); err != nil {
		t.Fatal(err)
	}
	if want, have := int32(1), atomic.LoadInt32(&(c.int32)); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

type countingCollector struct{ int32 }

func (c *countingCollector) Collect(*zipkin.Span) error { atomic.AddInt32(&(c.int32), 1); return nil }
