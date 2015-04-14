package zipkin_test

import (
	"math/rand"
	"net/http"
	"testing"

	"github.com/peterbourgon/gokit/tracing/zipkin"
	"golang.org/x/net/context"
)

func TestGeneration(t *testing.T) {
	rand.Seed(123)

	r, _ := http.NewRequest("GET", "http://cool.horse", nil)
	ctx := zipkin.GetFromHTTP(context.Background(), r)

	for key, want := range map[string]string{
		zipkin.TraceIDContextKey: "4a68998bed5c40f1",
		zipkin.SpanIDContextKey:  "35b51599210f9ba",
	} {
		val := ctx.Value(key)
		if val == nil {
			t.Errorf("%s: no entry", key)
			continue
		}
		have, ok := val.(string)
		if !ok {
			t.Errorf("%s: value not a string", key)
			continue
		}
		if want != have {
			t.Errorf("%s: want %q, have %q", key, want, have)
			continue
		}
	}
}

func TestHTTPHeaders(t *testing.T) {
	ids := map[string]string{
		zipkin.TraceIDContextKey:      "some_trace_id",
		zipkin.SpanIDContextKey:       "some_span_id",
		zipkin.ParentSpanIDContextKey: "some_parent_span_id",
	}

	ctx0 := context.Background()
	for key, val := range ids {
		ctx0 = context.WithValue(ctx0, key, val)
	}
	r, _ := http.NewRequest("GET", "http://best.horse", nil)
	zipkin.SetHTTPHeaders(ctx0, r.Header)
	ctx1 := zipkin.GetFromHTTP(context.Background(), r)

	for key, want := range ids {
		val := ctx1.Value(key)
		if val == nil {
			t.Errorf("%s: no entry", key)
			continue
		}
		have, ok := val.(string)
		if !ok {
			t.Errorf("%s: value not a string", key)
			continue
		}
		if want != have {
			t.Errorf("%s: want %q, have %q", key, want, have)
			continue
		}
	}
}
