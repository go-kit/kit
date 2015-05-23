package zipkin

import (
	"math/rand"
	"net/http"
	"strconv"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/server"

	"golang.org/x/net/context"
)

// In Zipkin, "spans are considered to start and stop with the client." The
// client is responsible for creating a new span ID for each outgoing request,
// copying its span ID to the parent span ID, and maintaining the same trace
// ID. The server-receive and server-send annotations can be considered value
// added information and aren't strictly necessary.
//
// Further reading:
// • http://www.slideshare.net/johanoskarsson/zipkin-runtime-open-house
// • https://groups.google.com/forum/#!topic/zipkin-user/KilwtSA0g1k
// • https://gist.github.com/yoavaa/3478d3a0df666f21a98c

const (
	// https://github.com/racker/tryfer#headers
	traceIDHTTPHeader      = "X-B3-TraceId"
	spanIDHTTPHeader       = "X-B3-SpanId"
	parentSpanIDHTTPHeader = "X-B3-ParentSpanId"

	// ClientSend is the annotation value used to mark a client sending a
	// request to a server.
	ClientSend = "cs"

	// ServerReceive is the annotation value used to mark a server's receipt
	// of a request from a client.
	ServerReceive = "sr"

	// ServerSend is the annotation value used to mark a server's completion
	// of a request and response to a client.
	ServerSend = "ss"

	// ClientReceive is the annotation value used to mark a client's receipt
	// of a completed request from a server.
	ClientReceive = "cr"
)

// AnnotateEndpoint extracts a span from the context, adds server-receive and
// server-send annotations at the boundaries, and submits the span to the
// collector. If no span is present, a new span is generated and put in the
// context.
func AnnotateEndpoint(newSpan NewSpanFunc, c Collector) func(server.Endpoint) server.Endpoint {
	return func(e server.Endpoint) server.Endpoint {
		return func(ctx context.Context, req server.Request) (server.Response, error) {
			span, ctx := mustGetServerSpan(ctx, newSpan)
			span.Annotate(ServerReceive)
			defer func() { span.Annotate(ServerSend); c.Collect(span) }()
			return e(ctx, req)
		}
	}
}

// FromHTTP is a helper method that allows NewSpanFunc's factory function to
// be easily invoked by passing an HTTP request. The span name is the HTTP
// method. The trace, span, and parent span IDs are taken from the request
// headers.
func FromHTTP(newSpan NewSpanFunc) func(*http.Request) *Span {
	return func(r *http.Request) *Span {
		traceIDStr := r.Header.Get(traceIDHTTPHeader)
		if traceIDStr == "" {
			// If there's no trace ID, that's normal: just make a new one.
			log.DefaultLogger.Log("debug", "make new span")
			return newSpan(newID(), newID(), 0)
		}
		traceID, err := strconv.ParseInt(traceIDStr, 16, 64)
		if err != nil {
			log.DefaultLogger.Log(traceIDHTTPHeader, traceIDStr, "err", err)
			return newSpan(newID(), newID(), 0)
		}
		spanIDStr := r.Header.Get(spanIDHTTPHeader)
		if spanIDStr == "" {
			log.DefaultLogger.Log("msg", "trace ID without span ID") // abnormal
			spanIDStr = strconv.FormatInt(newID(), 64)               // deal with it
		}
		spanID, err := strconv.ParseInt(spanIDStr, 16, 64)
		if err != nil {
			log.DefaultLogger.Log(spanIDHTTPHeader, spanIDStr, "err", err) // abnormal
			spanID = newID()                                               // deal with it
		}
		parentSpanIDStr := r.Header.Get(parentSpanIDHTTPHeader)
		if parentSpanIDStr == "" {
			parentSpanIDStr = "0" // normal
		}
		parentSpanID, err := strconv.ParseInt(parentSpanIDStr, 16, 64)
		if err != nil {
			log.DefaultLogger.Log(parentSpanIDHTTPHeader, parentSpanIDStr, "err", err) // abnormal
			parentSpanID = 0                                                           // the only way to deal with it
		}
		return newSpan(traceID, spanID, parentSpanID)
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
func NewChildSpan(ctx context.Context, newSpan NewSpanFunc) *Span {
	val := ctx.Value(SpanContextKey)
	if val == nil {
		return newSpan(newID(), newID(), 0)
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
	return newSpan(traceID, spanID, parentSpanID)
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

func mustGetServerSpan(ctx context.Context, newSpan NewSpanFunc) (*Span, context.Context) {
	val := ctx.Value(SpanContextKey)
	if val == nil {
		span := newSpan(newID(), newID(), 0)
		return span, context.WithValue(ctx, SpanContextKey, span)
	}
	span, ok := val.(*Span)
	if !ok {
		panic(SpanContextKey + " value isn't a span object")
	}
	return span, ctx
}

//func getID(h http.Header, key string) int64 {
//	val := h.Get(key)
//	if val == "" {
//		return 0
//	}
//	i, err := strconv.ParseInt(val, 16, 64)
//	if err != nil {
//		panic("invalid Zipkin ID in HTTP header: " + val)
//	}
//	return i
//}

func newID() int64 {
	// https://github.com/wadey/go-zipkin/blob/46e5f01/trace.go#L183-188
	// https://github.com/twitter/zipkin/issues/199
	// :(
	return rand.Int63() & 0x001fffffffffffff
}
