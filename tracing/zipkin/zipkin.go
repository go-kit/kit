package zipkin

import (
	"math/rand"
	"net/http"
	"strconv"

	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport/grpc"
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
	// SpanContextKey holds the key used to store Zipkin spans in the context.
	SpanContextKey = "Zipkin-Span"

	// https://github.com/racker/tryfer#headers
	traceIDHTTPHeader      = "X-B3-TraceId"
	spanIDHTTPHeader       = "X-B3-SpanId"
	parentSpanIDHTTPHeader = "X-B3-ParentSpanId"
	// gRPC keys are always lowercase
	traceIDGRPCKey      = "x-b3-traceid"
	spanIDGRPCKey       = "x-b3-spanid"
	parentSpanIDGRPCKey = "x-b3-parentspanid"

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

// AnnotateServer returns a server.Middleware that extracts a span from the
// context, adds server-receive and server-send annotations at the boundaries,
// and submits the span to the collector. If no span is found in the context,
// a new span is generated and inserted.
func AnnotateServer(newSpan NewSpanFunc, c Collector) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			span, ok := FromContext(ctx)
			if !ok {
				span = newSpan(newID(), newID(), 0)
				ctx = context.WithValue(ctx, SpanContextKey, span)
			}
			span.Annotate(ServerReceive)
			defer func() { span.Annotate(ServerSend); c.Collect(span) }()
			return next(ctx, request)
		}
	}
}

// AnnotateClient returns a middleware that extracts a parent span from the
// context, produces a client (child) span from it, adds client-send and
// client-receive annotations at the boundaries, and submits the span to the
// collector. If no span is found in the context, a new span is generated and
// inserted.
func AnnotateClient(newSpan NewSpanFunc, c Collector) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			var clientSpan *Span
			parentSpan, ok := FromContext(ctx)
			if ok {
				clientSpan = newSpan(parentSpan.TraceID(), newID(), parentSpan.SpanID())
			} else {
				clientSpan = newSpan(newID(), newID(), 0)
			}
			ctx = context.WithValue(ctx, SpanContextKey, clientSpan)                    // set
			defer func() { ctx = context.WithValue(ctx, SpanContextKey, parentSpan) }() // reset
			clientSpan.Annotate(ClientSend)
			defer func() { clientSpan.Annotate(ClientReceive); c.Collect(clientSpan) }()
			return next(ctx, request)
		}
	}
}

// ToContext returns a function that satisfies transport/http.BeforeFunc. It
// takes a Zipkin span from the incoming HTTP request, and saves it in the
// request context. It's designed to be wired into a server's HTTP transport
// Before stack. The logger is used to report errors.
func ToContext(newSpan NewSpanFunc, logger log.Logger) func(ctx context.Context, r *http.Request) context.Context {
	return func(ctx context.Context, r *http.Request) context.Context {
		return context.WithValue(ctx, SpanContextKey, fromHTTP(newSpan, r, logger))
	}
}

// ToGRPCContext returns a function that satisfies transport/grpc.BeforeFunc. It
// takes a Zipkin span from the incoming GRPC request, and saves it in the
// request context. It's designed to be wired into a server's GRPC transport
// Before stack. The logger is used to report errors.
func ToGRPCContext(newSpan NewSpanFunc, logger log.Logger) func(ctx context.Context, md *metadata.MD) context.Context {
	return func(ctx context.Context, md *metadata.MD) context.Context {
		return context.WithValue(ctx, SpanContextKey, fromGRPC(newSpan, *md, logger))
	}
}

// ToRequest returns a function that satisfies transport/http.BeforeFunc. It
// takes a Zipkin span from the context, and injects it into the HTTP request.
// It's designed to be wired into a client's HTTP transport Before stack. It's
// expected that AnnotateClient has already ensured the span in the context is
// a child/client span.
func ToRequest(newSpan NewSpanFunc) func(ctx context.Context, r *http.Request) context.Context {
	return func(ctx context.Context, r *http.Request) context.Context {
		span, ok := FromContext(ctx)
		if !ok {
			span = newSpan(newID(), newID(), 0)
		}
		if id := span.TraceID(); id > 0 {
			r.Header.Set(traceIDHTTPHeader, strconv.FormatInt(id, 16))
		}
		if id := span.SpanID(); id > 0 {
			r.Header.Set(spanIDHTTPHeader, strconv.FormatInt(id, 16))
		}
		if id := span.ParentSpanID(); id > 0 {
			r.Header.Set(parentSpanIDHTTPHeader, strconv.FormatInt(id, 16))
		}
		return ctx
	}
}

// ToGRPCRequest returns a function that satisfies transport/grpc.BeforeFunc. It
// takes a Zipkin span from the context, and injects it into the GRPC context.
// It's designed to be wired into a client's GRPC transport Before stack. It's
// expected that AnnotateClient has already ensured the span in the context is
// a child/client span.
func ToGRPCRequest(newSpan NewSpanFunc) func(ctx context.Context, md *metadata.MD) context.Context {
	return func(ctx context.Context, md *metadata.MD) context.Context {
		span, ok := FromContext(ctx)
		if !ok {
			span = newSpan(newID(), newID(), 0)
		}
		if id := span.TraceID(); id > 0 {
			key, value := grpc.EncodeKeyValue(traceIDGRPCKey, strconv.FormatInt(id, 16))
			(*md)[key] = append((*md)[key], value)
		}
		if id := span.SpanID(); id > 0 {
			key, value := grpc.EncodeKeyValue(spanIDGRPCKey, strconv.FormatInt(id, 16))
			(*md)[key] = append((*md)[key], value)
		}
		if id := span.ParentSpanID(); id > 0 {
			key, value := grpc.EncodeKeyValue(parentSpanIDGRPCKey, strconv.FormatInt(id, 16))
			(*md)[key] = append((*md)[key], value)
		}
		return ctx
	}
}

func fromHTTP(newSpan NewSpanFunc, r *http.Request, logger log.Logger) *Span {
	traceIDStr := r.Header.Get(traceIDHTTPHeader)
	if traceIDStr == "" {
		return newSpan(newID(), newID(), 0) // normal; just make a new one
	}
	traceID, err := strconv.ParseInt(traceIDStr, 16, 64)
	if err != nil {
		logger.Log(traceIDHTTPHeader, traceIDStr, "err", err)
		return newSpan(newID(), newID(), 0)
	}
	spanIDStr := r.Header.Get(spanIDHTTPHeader)
	if spanIDStr == "" {
		logger.Log("msg", "trace ID without span ID") // abnormal
		spanIDStr = strconv.FormatInt(newID(), 64)    // deal with it
	}
	spanID, err := strconv.ParseInt(spanIDStr, 16, 64)
	if err != nil {
		logger.Log(spanIDHTTPHeader, spanIDStr, "err", err) // abnormal
		spanID = newID()                                    // deal with it
	}
	parentSpanIDStr := r.Header.Get(parentSpanIDHTTPHeader)
	if parentSpanIDStr == "" {
		parentSpanIDStr = "0" // normal
	}
	parentSpanID, err := strconv.ParseInt(parentSpanIDStr, 16, 64)
	if err != nil {
		logger.Log(parentSpanIDHTTPHeader, parentSpanIDStr, "err", err) // abnormal
		parentSpanID = 0                                                // the only way to deal with it
	}
	return newSpan(traceID, spanID, parentSpanID)
}

func fromGRPC(newSpan NewSpanFunc, md metadata.MD, logger log.Logger) *Span {
	traceIDSlc := md[traceIDGRPCKey]
	pos := len(traceIDSlc) - 1
	if pos < 0 {
		return newSpan(newID(), newID(), 0) // normal; just make a new one
	}
	traceID, err := strconv.ParseInt(traceIDSlc[pos], 16, 64)
	if err != nil {
		logger.Log(traceIDHTTPHeader, traceIDSlc, "err", err)
		return newSpan(newID(), newID(), 0)
	}
	spanIDSlc := md[spanIDGRPCKey]
	pos = len(spanIDSlc) - 1
	if pos < 0 {
		spanIDSlc = make([]string, 1)
		pos = 0
	}
	if spanIDSlc[pos] == "" {
		logger.Log("msg", "trace ID without span ID")   // abnormal
		spanIDSlc[pos] = strconv.FormatInt(newID(), 64) // deal with it
	}
	spanID, err := strconv.ParseInt(spanIDSlc[pos], 16, 64)
	if err != nil {
		logger.Log(spanIDHTTPHeader, spanIDSlc, "err", err) // abnormal
		spanID = newID()                                    // deal with it
	}
	parentSpanIDSlc := md[parentSpanIDGRPCKey]
	pos = len(parentSpanIDSlc) - 1
	if pos < 0 {
		parentSpanIDSlc = make([]string, 1)
		pos = 0
	}
	if parentSpanIDSlc[pos] == "" {
		parentSpanIDSlc[pos] = "0" // normal
	}
	parentSpanID, err := strconv.ParseInt(parentSpanIDSlc[pos], 16, 64)
	if err != nil {
		logger.Log(parentSpanIDHTTPHeader, parentSpanIDSlc, "err", err) // abnormal
		parentSpanID = 0                                                // the only way to deal with it
	}
	return newSpan(traceID, spanID, parentSpanID)
}

// FromContext extracts an existing Zipkin span if it is stored in the provided
// context. If you add context.Context as the first parameter in your service
// methods you can annotate spans from within business logic. Typical use case
// is to AnnotateDuration on interaction with resources like databases.
func FromContext(ctx context.Context) (*Span, bool) {
	val := ctx.Value(SpanContextKey)
	if val == nil {
		return nil, false
	}
	span, ok := val.(*Span)
	if !ok {
		panic(SpanContextKey + " value isn't a span object")
	}
	return span, true
}

func newID() int64 {
	// https://github.com/wadey/go-zipkin/blob/46e5f01/trace.go#L183-188
	// https://github.com/twitter/zipkin/issues/199
	// :(
	return rand.Int63() & 0x001fffffffffffff
}
