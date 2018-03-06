package zipkin

import (
	"context"

	zipkin "github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/model"
	"github.com/openzipkin/zipkin-go/propagation/b3"
	"google.golang.org/grpc/metadata"

	"github.com/go-kit/kit/log"
)

// ContextToGRPC returns a grpc RequestFunc that injects a Zipkin Span found in
// `ctx` into the grpc Metadata. If no such Span can be found, the RequestFunc
// is a noop.
func ContextToGRPC(tracer *zipkin.Tracer, logger log.Logger) func(context.Context, *metadata.MD) context.Context {
	return func(ctx context.Context, md *metadata.MD) context.Context {
		if span := zipkin.SpanFromContext(ctx); span != nil {
			// There's nothing we can do with an error here.
			if err := b3.InjectGRPC(md)(span.Context()); err != nil {
				logger.Log("err", err)
			}
		}
		return ctx
	}
}

// GRPCToContext returns a grpc RequestFunc that tries to join with a Zipkin
// trace found in `req` and starts a new Span called `operationName`
// accordingly. If no trace could be found in `req`, the Span
// will be a trace root. The Span is incorporated in the returned Context and
// can be retrieved with zipkin.SpanFromContext(ctx).
func GRPCToContext(tracer *zipkin.Tracer, operationName string, logger log.Logger) func(ctx context.Context, md metadata.MD) context.Context {
	return func(ctx context.Context, md metadata.MD) context.Context {
		spanContext := tracer.Extract(b3.ExtractGRPC(&md))
		span := tracer.StartSpan(operationName, zipkin.Kind(model.Server), zipkin.Parent(spanContext))
		return zipkin.NewContext(ctx, span)
	}
}
