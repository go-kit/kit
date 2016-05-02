package opentracing

import (
	"net/http"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"
)

// ToHTTPRequest returns an http RequestFunc that injects an OpenTracing Span
// found in `ctx` into the http headers. If no such Span can be found, the
// RequestFunc is a noop.
func ToHTTPRequest(tracer opentracing.Tracer) kithttp.RequestFunc {
	return func(ctx context.Context, req *http.Request) context.Context {
		// Try to find a Span in the Context.
		if span := opentracing.SpanFromContext(ctx); span != nil {
			// There's nothing we can do with any errors here.
			_ = tracer.Inject(span, opentracing.TextMap, opentracing.HTTPHeaderTextMapCarrier(req.Header))
		}
		return ctx
	}
}

// FromHTTPRequest returns an http RequestFunc that tries to join with an
// OpenTracing trace found in `req` and starts a new Span called
// `operationName` accordingly. If no trace could be found in `req`, the Span
// will be a trace root. The Span is incorporated in the returned Context and
// can be retrieved with opentracing.SpanFromContext(ctx).
func FromHTTPRequest(tracer opentracing.Tracer, operationName string) kithttp.RequestFunc {
	return func(ctx context.Context, req *http.Request) context.Context {
		// Try to join to a trace propagated in `req`. There's nothing we can
		// do with any errors here, so we ignore them.
		span, _ := tracer.Join(operationName, opentracing.TextMap, opentracing.HTTPHeaderTextMapCarrier(req.Header))
		if span == nil {
			span = opentracing.StartSpan(operationName)
		}
		return opentracing.ContextWithSpan(ctx, span)
	}
}
