package zipkin

import (
	"math/rand"
	"net/http"
	"strconv"

	"golang.org/x/net/context"
)

// http://www.slideshare.net/johanoskarsson/zipkin-runtime-open-house

const (
	// https://github.com/racker/tryfer#headers
	traceIDHTTPHeader      = "X-B3-TraceId"
	spanIDHTTPHeader       = "X-B3-SpanId"
	parentSpanIDHTTPHeader = "X-B3-ParentSpanId"
)

// ViaHTTP is a helper method that allows NewSpanFunc's factory function to be
// easily invoked by passing an HTTP request. The span name is the HTTP
// method. The trace, span, and parent span IDs are taken from the request
// headers.
func ViaHTTP(f func(string, int64, int64, int64) *Span) func(*http.Request) *Span {
	return func(r *http.Request) *Span {
		return f(
			r.Method,
			getID(r.Header, traceIDHTTPHeader),
			getID(r.Header, spanIDHTTPHeader),
			getID(r.Header, parentSpanIDHTTPHeader),
		)
	}
}

// ToContext returns a function satisfies transport/http.BeforeFunc. When
// invoked, it generates a Zipkin span from the incoming HTTP request, and
// saves it in the request context under the SpanContextKey.
func ToContext(f func(*http.Request) *Span) func(context.Context, *http.Request) context.Context {
	return func(ctx context.Context, r *http.Request) context.Context {
		return context.WithValue(ctx, SpanContextKey, f(r))
	}
}

// SetRequestHeaders sets up HTTP headers for a new outgoing request, based on
// the Span in the request context. It transparently passes through the trace
// ID, assigns a new, random span ID, and sets the parent span ID to the
// current span ID. All IDs are encoded as hex strings.
//
// This function is meant to be applied to outgoing HTTP requests.
func SetRequestHeaders(ctx context.Context, h http.Header) error {
	val := ctx.Value(SpanContextKey)
	if val == nil {
		return ErrSpanNotFound
	}

	span, ok := val.(*Span)
	if !ok {
		panic(SpanContextKey + " value isn't a span object")
	}

	if id := span.TraceID(); id > 0 {
		h.Set(traceIDHTTPHeader, strconv.FormatInt(id, 16))
	}

	h.Set(spanIDHTTPHeader, strconv.FormatInt(newID(), 16))

	if id := span.SpanID(); id > 0 {
		h.Set(parentSpanIDHTTPHeader, strconv.FormatInt(id, 16))
	}

	return nil
}

func getID(h http.Header, key string) int64 {
	val := h.Get(key)
	if val == "" {
		return 0
	}
	i, err := strconv.ParseInt(val, 16, 64)
	if err != nil {
		panic("invalid Zipkin ID in HTTP header: " + val)
	}
	return i
}

func newID() int64 {
	// https://github.com/wadey/go-zipkin/blob/46e5f01/trace.go#L183-188
	// https://github.com/twitter/zipkin/issues/199
	// :(
	return rand.Int63() & 0x001fffffffffffff
}
