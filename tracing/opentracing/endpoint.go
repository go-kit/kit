package opentracing

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	"github.com/go-kit/kit/endpoint"
)

// TraceServer returns a Middleware that wraps the `next` Endpoint in an
// OpenTracing Span called `operationName`.
//
// If `ctx` already has a Span, it is re-used and the operation name is
// overwritten. If `ctx` does not yet have a Span, one is created here.
func TraceServer(tracer opentracing.Tracer, operationName string, opts ...EndpointOption) endpoint.Middleware {
	cfg := &EndpointOptions{}

	for _, opt := range opts {
		opt(cfg)
	}

	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			if cfg.GetOperationName != nil {
				if newOperationName := cfg.GetOperationName(ctx, operationName); newOperationName != "" {
					operationName = newOperationName
				}
			}

			serverSpan := opentracing.SpanFromContext(ctx)
			if serverSpan == nil {
				// All we can do is create a new root span.
				serverSpan = tracer.StartSpan(operationName)
			} else {
				serverSpan.SetOperationName(operationName)
			}
			defer serverSpan.Finish()
			ext.SpanKindRPCServer.Set(serverSpan)
			ctx = opentracing.ContextWithSpan(ctx, serverSpan)

			applyTags(serverSpan, cfg.Tags)
			if cfg.GetTags != nil {
				extraTags := cfg.GetTags(ctx)
				applyTags(serverSpan, extraTags)
			}

			response, err := next(ctx, request)
			if err := identifyError(response, err, cfg.IgnoreBusinessError); err != nil {
				ext.LogError(serverSpan, err)
			}

			return response, err
		}
	}
}

// TraceClient returns a Middleware that wraps the `next` Endpoint in an
// OpenTracing Span called `operationName`.
func TraceClient(tracer opentracing.Tracer, operationName string, opts ...EndpointOption) endpoint.Middleware {
	cfg := &EndpointOptions{}

	for _, opt := range opts {
		opt(cfg)
	}

	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			if cfg.GetOperationName != nil {
				if newOperationName := cfg.GetOperationName(ctx, operationName); newOperationName != "" {
					operationName = newOperationName
				}
			}

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
			ext.SpanKindRPCClient.Set(clientSpan)
			ctx = opentracing.ContextWithSpan(ctx, clientSpan)

			applyTags(clientSpan, cfg.Tags)
			if cfg.GetTags != nil {
				extraTags := cfg.GetTags(ctx)
				applyTags(clientSpan, extraTags)
			}

			response, err := next(ctx, request)
			if err := identifyError(response, err, cfg.IgnoreBusinessError); err != nil {
				ext.LogError(clientSpan, err)
			}

			return response, err
		}
	}
}

func applyTags(span opentracing.Span, tags opentracing.Tags) {
	for key, value := range tags {
		span.SetTag(key, value)
	}
}

func identifyError(response interface{}, err error, ignoreBusinessError bool) error {
	if err != nil {
		return err
	}

	if !ignoreBusinessError {
		if res, ok := response.(endpoint.Failer); ok {
			return res.Failed()
		}
	}

	return nil
}
