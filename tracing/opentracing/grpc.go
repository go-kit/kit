package opentracing

import (
	"encoding/base64"
	"strings"

	"github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"

	"github.com/go-kit/kit/log"
)

// ToGRPCRequest returns a grpc RequestFunc that injects an OpenTracing Span
// found in `ctx` into the grpc Metadata. If no such Span can be found, the
// RequestFunc is a noop.
//
// The logger is used to report errors and may be nil.
func ToGRPCRequest(tracer opentracing.Tracer, logger log.Logger) func(ctx context.Context, md *metadata.MD) context.Context {
	return func(ctx context.Context, md *metadata.MD) context.Context {
		if span := opentracing.SpanFromContext(ctx); span != nil {
			// There's nothing we can do with an error here.
			if err := tracer.Inject(span, opentracing.TextMap, metadataReaderWriter{md}); err != nil {
				logger.Log("err", err)
			}
		}
		return ctx
	}
}

// FromGRPCRequest returns a grpc RequestFunc that tries to join with an
// OpenTracing trace found in `req` and starts a new Span called
// `operationName` accordingly. If no trace could be found in `req`, the Span
// will be a trace root. The Span is incorporated in the returned Context and
// can be retrieved with opentracing.SpanFromContext(ctx).
//
// The logger is used to report errors and may be nil.
func FromGRPCRequest(tracer opentracing.Tracer, operationName string, logger log.Logger) func(ctx context.Context, md *metadata.MD) context.Context {
	return func(ctx context.Context, md *metadata.MD) context.Context {
		span, err := tracer.Join(operationName, opentracing.TextMap, metadataReaderWriter{md})
		if err != nil {
			span = tracer.StartSpan(operationName)
			if err != opentracing.ErrTraceNotFound {
				logger.Log("err", err)
			}
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
	key = strings.ToLower(key)
	if strings.HasSuffix(key, "-bin") {
		val = string(base64.StdEncoding.EncodeToString([]byte(val)))
	}
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
