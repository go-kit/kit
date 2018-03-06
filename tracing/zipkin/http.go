package zipkin

import (
	"context"
	"net/http"

	zipkin "github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/model"
	"github.com/openzipkin/zipkin-go/propagation/b3"

	"github.com/go-kit/kit/log"
	kithttp "github.com/go-kit/kit/transport/http"
)

// ContextToHTTP returns an http RequestFunc that injects a Zipkin Span found
// in `ctx` into the http headers. If no such Span can be found, the RequestFunc
// is a noop.
func ContextToHTTP(tracer *zipkin.Tracer, logger log.Logger) kithttp.RequestFunc {
	return func(ctx context.Context, req *http.Request) context.Context {
		if span := zipkin.SpanFromContext(ctx); span != nil {
			// add some common Zipkin Tags
			zipkin.TagHTTPMethod.Set(span, req.Method)
			zipkin.TagHTTPUrl.Set(span, req.URL.String())
			if endpoint, err := zipkin.NewEndpoint("", req.URL.Host); err == nil {
				span.SetRemoteEndpoint(endpoint)
			}
			// There's nothing we can do with any errors here.
			if err := b3.InjectHTTP(req)(span.Context()); err != nil {
				logger.Log("err", err)
			}
		}
		return ctx
	}
}

// HTTPToContext returns an http RequestFunc that tries to join with a Zipkin
// trace found in `req` and starts a new Span called  `operationName`
// accordingly. If no trace could be found in `req`, the Span will be a trace
// root. The Span is incorporated in the returned Context and can be retrieved
// with zipkin.SpanFromContext(ctx).
func HTTPToContext(tracer *zipkin.Tracer, operationName string, logger log.Logger) kithttp.RequestFunc {
	return func(ctx context.Context, req *http.Request) context.Context {
		spanContext := tracer.Extract(b3.ExtractHTTP(req))
		span := tracer.StartSpan(
			operationName, zipkin.Kind(model.Server), zipkin.Parent(spanContext),
		)
		// add some common Zipkin Tags
		zipkin.TagHTTPMethod.Set(span, req.Method)
		zipkin.TagHTTPUrl.Set(span, req.URL.String())
		return zipkin.NewContext(ctx, span)
	}
}
