package zipkin

import (
	"net/http"
	"strconv"
	"testing"

	"golang.org/x/net/context"
)

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

	sf := MakeNewSpanFunc(hostport, serviceName, methodName)
	cf := ToContext(sf)

	ctx := cf(context.Background(), r)
	val := ctx.Value(SpanContextKey)
	if val == nil {
		t.Fatalf("%s returned no value", SpanContextKey)
	}
	span, ok := val.(*Span)
	if !ok {
		t.Fatalf("%s was not a Span object", SpanContextKey)
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
	setRequestHeaders(r.Header, NewSpan(hostport, serviceName, methodName, traceID, spanID, parentSpanID))

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
