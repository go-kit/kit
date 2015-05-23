package zipkin_test

import (
	"math/rand"
	"net/http"
	"strconv"
	"sync/atomic"
	"testing"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/server"
	"github.com/go-kit/kit/tracing/zipkin"
)

func TestAnnotateEndpoint(t *testing.T) {
	const (
		hostport    = "1.2.3.4:1234"
		serviceName = "some-service"
		methodName  = "some-method"
	)

	f := zipkin.MakeNewSpanFunc(hostport, serviceName, methodName)
	c := &countingCollector{}

	var e server.Endpoint
	e = func(context.Context, server.Request) (server.Response, error) { return struct{}{}, nil }
	e = zipkin.AnnotateEndpoint(f, c)(e)

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

func TestFromHTTPToContext(t *testing.T) {
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

	sf := zipkin.MakeNewSpanFunc(hostport, serviceName, methodName)
	hf := zipkin.FromHTTP(sf)
	cf := zipkin.ToContext(hf)

	ctx := cf(context.Background(), r)
	val := ctx.Value(zipkin.SpanContextKey)
	if val == nil {
		t.Fatalf("%s returned no value", zipkin.SpanContextKey)
	}
	span, ok := val.(*zipkin.Span)
	if !ok {
		t.Fatalf("%s was not a Span object", zipkin.SpanContextKey)
	}

	if want, have := traceID, span.TraceID(); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	if want, have := spanID, span.SpanID(); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	if want, have := parentSpanID, span.ParentSpanID(); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestNewChildSpan(t *testing.T) {
	rand.Seed(123)

	const (
		hostport           = "1.2.1.2:1212"
		serviceName        = "my-service"
		methodName         = "my-method"
		traceID      int64 = 123
		spanID       int64 = 456
		parentSpanID int64 = 789
	)

	f := zipkin.MakeNewSpanFunc(hostport, serviceName, methodName)
	ctx := context.WithValue(context.Background(), zipkin.SpanContextKey, f(traceID, spanID, parentSpanID))
	childSpan := zipkin.NewChildSpan(ctx, f)

	if want, have := traceID, childSpan.TraceID(); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
	if have := childSpan.SpanID(); have == spanID {
		t.Errorf("span ID should be random, but we have %d", have)
	}
	if want, have := spanID, childSpan.ParentSpanID(); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestSetRequestHeaders(t *testing.T) {
	const (
		hostport           = "4.2.4.2:4242"
		serviceName        = "bar-service"
		methodName         = "bar-method"
		traceID      int64 = 123
		spanID       int64 = 456
		parentSpanID int64 = 789
	)

	r, _ := http.NewRequest("POST", "http://destroy.horse", nil)
	zipkin.SetRequestHeaders(r.Header, zipkin.NewSpan(hostport, serviceName, methodName, traceID, spanID, parentSpanID))

	for h, want := range map[string]string{
		"X-B3-TraceId":      strconv.FormatInt(traceID, 16),
		"X-B3-SpanId":       strconv.FormatInt(spanID, 16),
		"X-B3-ParentSpanId": strconv.FormatInt(parentSpanID, 16),
	} {
		if have := r.Header.Get(h); want != have {
			t.Errorf("%s: want %s, have %s", h, want, have)
		}
	}
}

type countingCollector struct{ int32 }

func (c *countingCollector) Collect(*zipkin.Span) error { atomic.AddInt32(&(c.int32), 1); return nil }
