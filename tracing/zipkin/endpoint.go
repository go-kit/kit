package zipkin

import (
	"context"

	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/model"

	"github.com/go-kit/kit/endpoint"
)

// TraceServer returns a Middleware that wraps the `next` Endpoint in a Zipkin
// Span called `operationName`.
//
// If `ctx` already has a Span, it is re-used and the operation name is
// overwritten. If `ctx` does not yet have a Span, one is created here.
func TraceServer(tracer *zipkin.Tracer, operationName string) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			var sp zipkin.Span
			// try to retrieve Span from Go context, create new Span if not found.
			if sp = zipkin.SpanFromContext(ctx); sp == nil {
				sp = tracer.StartSpan(operationName, zipkin.Kind(model.Server))
				ctx = zipkin.NewContext(ctx, sp)
			} else {
				sp.SetName(operationName)
			}
			defer sp.Finish()
			return next(ctx, request)
		}
	}
}

// TraceClient returns a Middleware that wraps the `next` Endpoint in a Zipkin
// Span called `operationName`.
func TraceClient(tracer *zipkin.Tracer, operationName string) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			var spanOpts = []zipkin.SpanOption{zipkin.Kind(model.Client)}
			// try to retrieve Span from Go context, use its SpanContext if found.
			if parentSpan := zipkin.SpanFromContext(ctx); parentSpan != nil {
				spanOpts = append(spanOpts, zipkin.Parent(parentSpan.Context()))
			}
			// create new client span (if sc is empty, Parent is a noop)
			sp := tracer.StartSpan(operationName, spanOpts...)
			defer sp.Finish()
			ctx = zipkin.NewContext(ctx, sp)
			return next(ctx, request)
		}
	}
}
