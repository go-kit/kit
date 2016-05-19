package opentracing

import (
	"github.com/opentracing/opentracing-go"
	otext "github.com/opentracing/opentracing-go/ext"
	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
)

// TraceServer returns a Middleware that wraps the `next` Endpoint in an
// OpenTracing Span called `operationName`.
//
// If `ctx` already has a Span, it is re-used and the operation name is
// overwritten. If `ctx` does not yet have a Span, one is created here.
func TraceServer(tracer opentracing.Tracer, operationName string) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			serverSpan := opentracing.SpanFromContext(ctx)
			if serverSpan == nil {
				// All we can do is create a new root span.
				serverSpan = tracer.StartSpan(operationName)
			} else {
				serverSpan.SetOperationName(operationName)
			}
			defer serverSpan.Finish()
			otext.SpanKind.Set(serverSpan, otext.SpanKindRPCServer)
			ctx = opentracing.ContextWithSpan(ctx, serverSpan)
			return next(ctx, request)
		}
	}
}

// TraceClient returns a Middleware that wraps the `next` Endpoint in an
// OpenTracing Span called `operationName`.
func TraceClient(tracer opentracing.Tracer, operationName string) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			parentSpan := opentracing.SpanFromContext(ctx)
			clientSpan := tracer.StartSpanWithOptions(opentracing.StartSpanOptions{
				OperationName: operationName,
				Parent:        parentSpan, // may be nil
			})
			defer clientSpan.Finish()
			otext.SpanKind.Set(clientSpan, otext.SpanKindRPCClient)
			ctx = opentracing.ContextWithSpan(ctx, clientSpan)
			return next(ctx, request)
		}
	}
}
