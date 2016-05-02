package opentracing

import (
	"github.com/go-kit/kit/endpoint"
	"github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"
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
			serverSpan.SetTag("span.kind", "server")
			ctx = opentracing.ContextWithSpan(ctx, serverSpan)
			defer serverSpan.Finish()
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
			clientSpan.SetTag("span.kind", "client")
			ctx = opentracing.ContextWithSpan(ctx, clientSpan)
			defer clientSpan.Finish()
			return next(ctx, request)
		}
	}
}
