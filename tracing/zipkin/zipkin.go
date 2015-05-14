package zipkin

import (
	"math/rand"
	"net/http"
	"strconv"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/server"
)

// http://www.slideshare.net/johanoskarsson/zipkin-runtime-open-house
// https://groups.google.com/forum/#!topic/zipkin-user/KilwtSA0g1k
// https://gist.github.com/yoavaa/3478d3a0df666f21a98c

const (
	// https://github.com/racker/tryfer#headers
	traceIDHTTPHeader      = "X-B3-TraceId"
	spanIDHTTPHeader       = "X-B3-SpanId"
	parentSpanIDHTTPHeader = "X-B3-ParentSpanId"

	clientSend    = "cs"
	serverReceive = "sr"
	serverSend    = "ss"
	clientReceive = "cr"
)

// AnnotateEndpoint extracts a span from the context, adds server-receive and
// server-send annotations at the boundaries, and submits the span to the
// collector. If no span is present, a new span is generated and put in the
// context.
func AnnotateEndpoint(f func(int64, int64, int64) *Span, c Collector) func(server.Endpoint) server.Endpoint {
	return func(e server.Endpoint) server.Endpoint {
		return func(ctx context.Context, req server.Request) (server.Response, error) {
			span, ctx := mustGetServerSpan(ctx, f)
			span.Annotate(serverReceive)
			defer func() { span.Annotate(serverSend); c.Collect(span) }()
			return e(ctx, req)
		}
	}
}

// FromHTTP is a helper method that allows NewSpanFunc's factory function to
// be easily invoked by passing an HTTP request. The span name is the HTTP
// method. The trace, span, and parent span IDs are taken from the request
// headers.
func FromHTTP(f func(int64, int64, int64) *Span) func(*http.Request) *Span {
	return func(r *http.Request) *Span {
		return f(
			getID(r.Header, traceIDHTTPHeader),
			getID(r.Header, spanIDHTTPHeader),
			getID(r.Header, parentSpanIDHTTPHeader),
		)
	}
}

// ToContext returns a function that satisfies transport/http.BeforeFunc. When
// invoked, it generates a Zipkin span from the incoming HTTP request, and
// saves it in the request context under the SpanContextKey.
func ToContext(f func(*http.Request) *Span) func(context.Context, *http.Request) context.Context {
	return func(ctx context.Context, r *http.Request) context.Context {
		return context.WithValue(ctx, SpanContextKey, f(r))
	}
}

// NewChildSpan creates a new child (client) span. If a span is present in the
// context, it will be interpreted as the parent.
func NewChildSpan(ctx context.Context, f func(int64, int64, int64) *Span) *Span {
	val := ctx.Value(SpanContextKey)
	if val == nil {
		return f(newID(), newID(), 0)
	}
	parentSpan, ok := val.(*Span)
	if !ok {
		panic(SpanContextKey + " value isn't a span object")
	}
	var (
		traceID      = parentSpan.TraceID()
		spanID       = newID()
		parentSpanID = parentSpan.SpanID()
	)
	return f(traceID, spanID, parentSpanID)
}

// SetRequestHeaders sets up HTTP headers for a new outbound request based on
// the (client) span. All IDs are encoded as hex strings.
func SetRequestHeaders(h http.Header, s *Span) {
	if id := s.TraceID(); id > 0 {
		h.Set(traceIDHTTPHeader, strconv.FormatInt(id, 16))
	}
	if id := s.SpanID(); id > 0 {
		h.Set(spanIDHTTPHeader, strconv.FormatInt(id, 16))
	}
	if id := s.ParentSpanID(); id > 0 {
		h.Set(parentSpanIDHTTPHeader, strconv.FormatInt(id, 16))
	}
}

func mustGetServerSpan(ctx context.Context, f func(int64, int64, int64) *Span) (*Span, context.Context) {
	val := ctx.Value(SpanContextKey)
	if val == nil {
		span := f(newID(), newID(), 0)
		return span, context.WithValue(ctx, SpanContextKey, span)
	}
	span, ok := val.(*Span)
	if !ok {
		panic(SpanContextKey + " value isn't a span object")
	}
	return span, ctx
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
