package zipkin

import (
	"context"
	"strconv"

	zipkin "github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/model"
	"github.com/openzipkin/zipkin-go/propagation/b3"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/go-kit/kit/log"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
)

// GRPCClientTrace enables Zipkin tracing of a Go kit gRPC Client Transport.
func GRPCClientTrace(tracer *zipkin.Tracer, options ...Option) kitgrpc.ClientOption {
	config := tracerOptions{
		tags:      make(map[string]string),
		name:      "",
		logger:    log.NewNopLogger(),
		propagate: true,
	}

	for _, option := range options {
		option(&config)
	}

	clientBefore := kitgrpc.ClientBefore(
		func(ctx context.Context, md *metadata.MD) context.Context {
			var (
				spanContext model.SpanContext
				name        string
			)

			if config.name != "" {
				name = config.name
			} else {
				name = ctx.Value(kitgrpc.ContextKeyRequestMethod).(string)
			}

			if parent := zipkin.SpanFromContext(ctx); parent != nil {
				spanContext = parent.Context()
			}

			span := tracer.StartSpan(
				name,
				zipkin.Kind(model.Client),
				zipkin.Tags(config.tags),
				zipkin.Parent(spanContext),
				zipkin.FlushOnFinish(false),
			)

			if config.propagate {
				if err := b3.InjectGRPC(md)(span.Context()); err != nil {
					config.logger.Log("err", err)
				}
			}

			return zipkin.NewContext(ctx, span)
		},
	)

	clientAfter := kitgrpc.ClientAfter(
		func(ctx context.Context, _ metadata.MD, _ metadata.MD) context.Context {
			if span := zipkin.SpanFromContext(ctx); span != nil {
				span.Finish()
			}

			return ctx
		},
	)

	clientFinalizer := kitgrpc.ClientFinalizer(
		func(ctx context.Context, err error) {
			if span := zipkin.SpanFromContext(ctx); span != nil {
				if err != nil {
					zipkin.TagError.Set(span, err.Error())
				}
				// calling span.Finish() a second time is a noop, if we didn't get to
				// ClientAfter we can at least time the early bail out by calling it
				// here.
				span.Finish()
				// send span to the Reporter
				span.Flush()
			}
		},
	)

	return func(c *kitgrpc.Client) {
		clientBefore(c)
		clientAfter(c)
		clientFinalizer(c)
	}

}

// GRPCServerTrace enables Zipkin tracing of a Go kit gRPC Server Transport.
func GRPCServerTrace(tracer *zipkin.Tracer, options ...Option) kitgrpc.ServerOption {
	config := tracerOptions{
		tags:      make(map[string]string),
		name:      "",
		logger:    log.NewNopLogger(),
		propagate: true,
	}

	for _, option := range options {
		option(&config)
	}

	serverBefore := kitgrpc.ServerBefore(
		func(ctx context.Context, md metadata.MD) context.Context {
			var (
				spanContext model.SpanContext
				name        string
				tags        = make(map[string]string)
			)

			rpcMethod, ok := ctx.Value(kitgrpc.ContextKeyRequestMethod).(string)
			if !ok {
				config.logger.Log("unable to retrieve method name: missing gRPC interceptor hook")
			} else {
				tags["grpc.method"] = rpcMethod
			}

			if config.name != "" {
				name = config.name
			} else {
				name = rpcMethod
			}

			if config.propagate {
				spanContext = tracer.Extract(b3.ExtractGRPC(&md))
				if spanContext.Err != nil {
					config.logger.Log("err", spanContext.Err)
				}
			}

			span := tracer.StartSpan(
				name,
				zipkin.Kind(model.Server),
				zipkin.Tags(config.tags),
				zipkin.Tags(tags),
				zipkin.Parent(spanContext),
				zipkin.FlushOnFinish(false),
			)

			return zipkin.NewContext(ctx, span)
		},
	)

	serverAfter := kitgrpc.ServerAfter(
		func(ctx context.Context, _ *metadata.MD, _ *metadata.MD) context.Context {
			if span := zipkin.SpanFromContext(ctx); span != nil {
				span.Finish()
			}

			return ctx
		},
	)

	serverFinalizer := kitgrpc.ServerFinalizer(
		func(ctx context.Context, err error) {
			if span := zipkin.SpanFromContext(ctx); span != nil {
				if err != nil {
					if status, ok := status.FromError(err); ok {
						statusCode := strconv.FormatUint(uint64(status.Code()), 10)
						zipkin.TagGRPCStatusCode.Set(span, statusCode)
						zipkin.TagError.Set(span, status.Message())
					} else {
						zipkin.TagError.Set(span, err.Error())
					}
				}

				// calling span.Finish() a second time is a noop, if we didn't get to
				// ServerAfter we can at least time the early bail out by calling it
				// here.
				span.Finish()
				// send span to the Reporter
				span.Flush()
			}
		},
	)

	return func(s *kitgrpc.Server) {
		serverBefore(s)
		serverAfter(s)
		serverFinalizer(s)
	}
}
