package opentracing

import (
	"context"

	"github.com/opentracing/opentracing-go"
	otext "github.com/opentracing/opentracing-go/ext"

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
			otext.SpanKindRPCServer.Set(serverSpan)
			ctx = opentracing.ContextWithSpan(ctx, serverSpan)

			response, err := next(ctx, request)

			spanErr := effectiveSpanError(response, err)
			if spanErr != nil {
				markSpanWithError(serverSpan, spanErr)
			}

			return response, err
		}
	}
}

// TraceClient returns a Middleware that wraps the `next` Endpoint in an
// OpenTracing Span called `operationName`.
func TraceClient(tracer opentracing.Tracer, operationName string) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			var clientSpan opentracing.Span
			if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
				clientSpan = tracer.StartSpan(
					operationName,
					opentracing.ChildOf(parentSpan.Context()),
				)
			} else {
				clientSpan = tracer.StartSpan(operationName)
			}
			defer clientSpan.Finish()
			otext.SpanKindRPCClient.Set(clientSpan)
			ctx = opentracing.ContextWithSpan(ctx, clientSpan)

			response, err := next(ctx, request)

			spanErr := effectiveSpanError(response, err)
			if spanErr != nil {
				markSpanWithError(clientSpan, spanErr)
			}

			return response, err
		}
	}
}

func effectiveSpanError(endpointResp interface{}, endpointErr error) error {
	if endpointErr != nil {
		return endpointErr
	}

	if res, ok := endpointResp.(endpoint.Failer); ok && res.Failed() != nil {
		return res.Failed()
	}

	return nil
}

func markSpanWithError(span opentracing.Span, err error) {
	otext.Error.Set(span, true)
	span.LogKV("error.object", err)
}
