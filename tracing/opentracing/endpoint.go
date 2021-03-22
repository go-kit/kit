package opentracing

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	"github.com/go-kit/kit/endpoint"
)

// TraceEndpoint returns a Middleware that wraps the `next` Endpoint in an
// OpenTracing Span called `operationName`.
//
// If `ctx` already has a Span, child span is created from it.
// If `ctx` doesn't yet have a Span, the new one is created.
func TraceEndpoint(tracer opentracing.Tracer, operationName string, opts ...EndpointOption) endpoint.Middleware {
	cfg := &EndpointOptions{
		Tags: make(opentracing.Tags),
	}

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

			span := opentracing.SpanFromContext(ctx)
			if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
				span = tracer.StartSpan(
					operationName,
					opentracing.ChildOf(parentSpan.Context()),
				)
			} else {
				span = tracer.StartSpan(operationName)
			}
			defer span.Finish()
			ctx = opentracing.ContextWithSpan(ctx, span)

			applyTags(span, cfg.Tags)
			if cfg.GetTags != nil {
				extraTags := cfg.GetTags(ctx)
				applyTags(span, extraTags)
			}

			response, err := next(ctx, request)
			if err := identifyError(response, err, cfg.IgnoreBusinessError); err != nil {
				ext.LogError(span, err)
			}

			return response, err
		}
	}
}

// TraceServer returns a Middleware that wraps the `next` Endpoint in an
// OpenTracing Span called `operationName` with server span.kind tag..
func TraceServer(tracer opentracing.Tracer, operationName string, opts ...EndpointOption) endpoint.Middleware {
	opts = append(opts, WithTags(map[string]interface{}{
		ext.SpanKindRPCServer.Key: ext.SpanKindRPCServer.Value,
	}))

	return TraceEndpoint(tracer, operationName, opts...)
}

// TraceClient returns a Middleware that wraps the `next` Endpoint in an
// OpenTracing Span called `operationName` with client span.kind tag.
func TraceClient(tracer opentracing.Tracer, operationName string, opts ...EndpointOption) endpoint.Middleware {
	opts = append(opts, WithTags(map[string]interface{}{
		ext.SpanKindRPCServer.Key: ext.SpanKindRPCClient.Value,
	}))

	return TraceEndpoint(tracer, operationName, opts...)
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
