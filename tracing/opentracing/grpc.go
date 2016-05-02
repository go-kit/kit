package opentracing

import (
	"github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

// ToGRPCRequest returns a grpc RequestFunc that injects an OpenTracing Span
// found in `ctx` into the grpc Metadata. If no such Span can be found, the
// RequestFunc is a noop.
func ToGRPCRequest(tracer opentracing.Tracer) func(ctx context.Context, md *metadata.MD) context.Context {
	return func(ctx context.Context, md *metadata.MD) context.Context {
		if span := opentracing.SpanFromContext(ctx); span != nil {
			// There's nothing we can do with an error here.
			_ = tracer.Inject(span, opentracing.TextMap, metadataReaderWriter{md})
		}
		return ctx
	}
}

// FromGRPCRequest returns a grpc RequestFunc that tries to join with an
// OpenTracing trace found in `req` and starts a new Span called
// `operationName` accordingly. If no trace could be found in `req`, the Span
// will be a trace root. The Span is incorporated in the returned Context and
// can be retrieved with opentracing.SpanFromContext(ctx).
func FromGRPCRequest(tracer opentracing.Tracer, operationName string) func(ctx context.Context, md *metadata.MD) context.Context {
	return func(ctx context.Context, md *metadata.MD) context.Context {
		span, err := tracer.Join(operationName, opentracing.TextMap, metadataReaderWriter{md})
		if err != nil || span == nil {
			span = tracer.StartSpan(operationName)
		}
		return opentracing.ContextWithSpan(ctx, span)
	}
}

// A type that conforms to opentracing.TextMapReader and
// opentracing.TextMapWriter.
type metadataReaderWriter struct {
	*metadata.MD
}

func (w metadataReaderWriter) Set(key, val string) {
	(*w.MD)[key] = append((*w.MD)[key], val)
}

func (w metadataReaderWriter) ForeachKey(handler func(key, val string) error) error {
	for k, vals := range *w.MD {
		for _, v := range vals {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}
