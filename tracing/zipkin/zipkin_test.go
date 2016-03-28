package zipkin_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
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
	toContext := zipkin.ToContext(newSpan, log.NewLogfmtLogger(ioutil.Discard))

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

func TestFromContext(t *testing.T) {
	const (
		hostport           = "5.5.5.5:5555"
		serviceName        = "foo-service"
		methodName         = "foo-method"
		traceID      int64 = 14
		spanID       int64 = 36
		parentSpanID int64 = 58
	)

	ctx := context.WithValue(
		context.Background(),
		zipkin.SpanContextKey,
		zipkin.NewSpan(hostport, serviceName, methodName, traceID, spanID, parentSpanID),
	)

	span, ok := zipkin.FromContext(ctx)
	if !ok {
		t.Fatalf("expected a context value in %q", zipkin.SpanContextKey)
	}
	if span == nil {
		t.Fatal("expected a Zipkin span object")
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

func TestToGRPCContext(t *testing.T) {
	const (
		hostport           = "5.5.5.5:5555"
		serviceName        = "foo-service"
		methodName         = "foo-method"
		traceID      int64 = 12
		spanID       int64 = 34
		parentSpanID int64 = 56
	)

	md := metadata.MD{
		"x-b3-traceid":      []string{strconv.FormatInt(traceID, 16)},
		"x-b3-spanid":       []string{strconv.FormatInt(spanID, 16)},
		"x-b3-parentspanid": []string{strconv.FormatInt(parentSpanID, 16)},
	}

	newSpan := zipkin.MakeNewSpanFunc(hostport, serviceName, methodName)
	toContext := zipkin.ToGRPCContext(newSpan, log.NewLogfmtLogger(ioutil.Discard))

	ctx := toContext(context.Background(), &md)
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

func TestToGRPCRequest(t *testing.T) {
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
	md := &metadata.MD{}
	ctx = zipkin.ToGRPCRequest(newSpan)(ctx, md)

	for header, wantInt := range map[string]int64{
		"x-b3-traceid":      traceID,
		"x-b3-spanid":       spanID,
		"x-b3-parentspanid": parentSpanID,
	} {
		if want, have := strconv.FormatInt(wantInt, 16), (*md)[header][0]; want != have {
			t.Errorf("%s: want %q, have %q", header, want, have)
		}
	}
}

func TestAnnotateServer(t *testing.T) {
	if err := testAnnotate(zipkin.AnnotateServer, zipkin.ServerReceive, zipkin.ServerSend); err != nil {
		t.Fatal(err)
	}
}

func TestAnnotateClient(t *testing.T) {
	if err := testAnnotate(zipkin.AnnotateClient, zipkin.ClientSend, zipkin.ClientReceive); err != nil {
		t.Fatal(err)
	}
}

func testAnnotate(
	annotate func(newSpan zipkin.NewSpanFunc, c zipkin.Collector) endpoint.Middleware,
	wantAnnotations ...string,
) error {
	const (
		hostport    = "1.2.3.4:1234"
		serviceName = "some-service"
		methodName  = "some-method"
	)

	newSpan := zipkin.MakeNewSpanFunc(hostport, serviceName, methodName)
	collector := &countingCollector{}

	var e endpoint.Endpoint
	e = func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil }
	e = annotate(newSpan, collector)(e)

	if want, have := 0, len(collector.annotations); want != have {
		return fmt.Errorf("pre-invocation: want %d, have %d", want, have)
	}
	if _, err := e(context.Background(), struct{}{}); err != nil {
		return fmt.Errorf("during invocation: %v", err)
	}
	if want, have := wantAnnotations, collector.annotations; !reflect.DeepEqual(want, have) {
		return fmt.Errorf("after invocation: want %v, have %v", want, have)
	}

	return nil
}

type countingCollector struct{ annotations []string }

func (c *countingCollector) Collect(s *zipkin.Span) error {
	for _, annotation := range s.Encode().GetAnnotations() {
		c.annotations = append(c.annotations, annotation.GetValue())
	}
	return nil
}

func (c *countingCollector) Close() error {
	return nil
}
