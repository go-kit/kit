package zipkin

import (
	"errors"
	"net/http"

	"github.com/peterbourgon/gokit/server"

	"golang.org/x/net/context"
)

const (
	// https://github.com/racker/tryfer#headers
	traceIDHTTPHeader      = "X-B3-TraceId"
	spanIDHTTPHeader       = "X-B3-SpanId"
	parentSpanIDHTTPHeader = "X-B3-ParentSpanId"

	// TraceIDContextKey holds the Zipkin TraceId, if available.
	TraceIDContextKey = "Zipkin-Trace-ID"

	// SpanIDContextKey holds the Zipkin SpanId, if available.
	SpanIDContextKey = "Zipkin-Span-ID"

	// ParentSpanIDContextKey holds the Zipkin ParentSpanId, if available.
	ParentSpanIDContextKey = "Zipkin-Parent-Span-ID"
)

// ErrMissingZipkinHeaders is returned when a context doesn't contain Zipkin
// trace, span, or parent span IDs.
var ErrMissingZipkinHeaders = errors.New("Zipkin headers missing from request context")

// GetHeaders extracts Zipkin headers from the HTTP request, and populates
// them into the context, if present.
func GetHeaders(ctx context.Context, header http.Header) context.Context {
	if val := header.Get(traceIDHTTPHeader); val != "" {
		ctx = context.WithValue(ctx, TraceIDContextKey, val)
	}
	if val := header.Get(spanIDHTTPHeader); val != "" {
		ctx = context.WithValue(ctx, SpanIDContextKey, val)
	}
	if val := header.Get(parentSpanIDHTTPHeader); val != "" {
		ctx = context.WithValue(ctx, ParentSpanIDContextKey, val)
	}
	return ctx
}

// RequireInContext implements the server.Gate allow func by checking if the
// context contains extracted Zipkin headers. Contexts without all headers
// aren't allowed to proceed.
func RequireInContext(ctx context.Context, _ server.Request) error {
	if ctx.Value(TraceIDContextKey) == nil || ctx.Value(SpanIDContextKey) == nil || ctx.Value(ParentSpanIDContextKey) == nil {
		return ErrMissingZipkinHeaders
	}
	return nil
}
