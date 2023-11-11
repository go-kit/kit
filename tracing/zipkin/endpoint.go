package zipkin

import (
	"context"

	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/model"

	"github.com/openmesh/kit/endpoint"
)

// TraceEndpoint returns an Endpoint middleware, tracing a Go kit endpoint.
// This endpoint tracer should be used in combination with a Go kit Transport
// tracing middleware or custom before and after transport functions as
// propagation of SpanContext is not provided in this middleware.
func TraceEndpoint[Request, Response any](tracer *zipkin.Tracer, name string) endpoint.Middleware[Request, Response] {
	return func(next endpoint.Endpoint[Request, Response]) endpoint.Endpoint[Request, Response] {
		return func(ctx context.Context, request Request) (Response, error) {
			var sc model.SpanContext
			if parentSpan := zipkin.SpanFromContext(ctx); parentSpan != nil {
				sc = parentSpan.Context()
			}
			sp := tracer.StartSpan(name, zipkin.Parent(sc))
			defer sp.Finish()

			ctx = zipkin.NewContext(ctx, sp)
			return next(ctx, request)
		}
	}
}
