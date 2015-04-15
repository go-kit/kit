package zipkin_test

import (
	"net/http"
	"strconv"
	"testing"

	"golang.org/x/net/context"

	"github.com/peterbourgon/gokit/tracing/zipkin"
)

func TestContextInjection(t *testing.T) {
	const (
		traceID      int64 = 12
		spanID       int64 = 34
		parentSpanID int64 = 56
	)

	r, _ := http.NewRequest("GET", "https://best.horse", nil)
	r.Header.Set("X-B3-TraceId", strconv.FormatInt(traceID, 16))
	r.Header.Set("X-B3-SpanId", strconv.FormatInt(spanID, 16))
	r.Header.Set("X-B3-ParentSpanId", strconv.FormatInt(parentSpanID, 16))

	sf := zipkin.NewSpanFunc("my-host", zipkin.NopCollector{})
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

func TestSetRequestHeaders(t *testing.T) {
	const (
		host               = "my-host"
		name               = "my-name"
		traceID      int64 = 123
		spanID       int64 = 456
		parentSpanID int64 = 789
	)

	span := zipkin.NewSpan(host, zipkin.NopCollector{}, name, traceID, spanID, parentSpanID)
	ctx := context.WithValue(context.Background(), zipkin.SpanContextKey, span)

	r, _ := http.NewRequest("POST", "http://destroy.horse", nil)
	if err := zipkin.SetRequestHeaders(ctx, r.Header); err != nil {
		t.Fatal(err)
	}

	for h, want := range map[string]string{
		"X-B3-TraceId": strconv.FormatInt(traceID, 16),
		// span ID is now random
		"X-B3-ParentSpanId": strconv.FormatInt(spanID, 16),
	} {
		if have := r.Header.Get(h); want != have {
			t.Errorf("%s: want %s, have %s", h, want, have)
		}
	}
}
