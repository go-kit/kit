package zipkin

import (
	"math/rand"
	"net/http"
	"strconv"

	"golang.org/x/net/context"
)

const (
	// https://github.com/racker/tryfer#headers
	traceIDHTTPHeader      = "X-B3-TraceId"
	spanIDHTTPHeader       = "X-B3-SpanId"
	parentSpanIDHTTPHeader = "X-B3-ParentSpanId"

	// TraceIDContextKey holds the Zipkin TraceId.
	TraceIDContextKey = "Zipkin-Trace-ID"

	// SpanIDContextKey holds the Zipkin SpanId.
	SpanIDContextKey = "Zipkin-Span-ID"

	// ParentSpanIDContextKey holds the Zipkin ParentSpanId, if available.
	ParentSpanIDContextKey = "Zipkin-Parent-Span-ID"
)

// GetFromHTTP implements transport/http.BeforeFunc, populating Zipkin headers
// into the context from the HTTP headers. It will generate new trace and span
// IDs if none are found.
func GetFromHTTP(ctx context.Context, r *http.Request) context.Context {
	if val := r.Header.Get(traceIDHTTPHeader); val != "" {
		ctx = context.WithValue(ctx, TraceIDContextKey, val)
	} else {
		ctx = context.WithValue(ctx, TraceIDContextKey, strconv.FormatInt(rand.Int63(), 16))
	}
	if val := r.Header.Get(spanIDHTTPHeader); val != "" {
		ctx = context.WithValue(ctx, SpanIDContextKey, val)
	} else {
		ctx = context.WithValue(ctx, SpanIDContextKey, strconv.FormatInt(rand.Int63(), 16))
	}
	if val := r.Header.Get(parentSpanIDHTTPHeader); val != "" {
		ctx = context.WithValue(ctx, ParentSpanIDContextKey, val)
	}
	return ctx
}

// SetHTTPHeaders copies Zipkin headers from the context into the HTTP header.
func SetHTTPHeaders(ctx context.Context, h http.Header) {
	for ctxKey, hdrKey := range map[string]string{
		TraceIDContextKey:      traceIDHTTPHeader,
		SpanIDContextKey:       spanIDHTTPHeader,
		ParentSpanIDContextKey: parentSpanIDHTTPHeader,
	} {
		if val := ctx.Value(ctxKey); val != nil {
			s, ok := val.(string)
			if !ok {
				panic("context value for " + ctxKey + " isn't string")
			}
			h.Set(hdrKey, s)
		}
	}
}
