package opencensus

import (
	"context"

	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	kitgrpc "github.com/go-kit/kit/transport/grpc"
)

const propagationKey = "grpc-trace-bin"

// GRPCClientTrace enables OpenCensus tracing of a Go kit gRPC transport client.
func GRPCClientTrace(options ...TracerOption) kitgrpc.ClientOption {
	cfg := TracerOptions{
		sampler: trace.AlwaysSample(),
	}

	for _, option := range options {
		option(&cfg)
	}

	clientBefore := kitgrpc.ClientBefore(
		func(ctx context.Context, md *metadata.MD) context.Context {
			var name string

			if cfg.name != "" {
				name = cfg.name
			} else {
				name = ctx.Value(kitgrpc.ContextKeyRequestMethod).(string)
			}

			span := trace.NewSpan(
				name,
				trace.FromContext(ctx),
				trace.StartOptions{
					Sampler:  cfg.sampler,
					SpanKind: trace.SpanKindClient,
				},
			)

			if !cfg.public {
				traceContextBinary := string(propagation.Binary(span.SpanContext()))
				(*md)[propagationKey] = append((*md)[propagationKey], traceContextBinary)
			}

			return trace.NewContext(ctx, span)
		},
	)

	clientFinalizer := kitgrpc.ClientFinalizer(
		func(ctx context.Context, err error) {
			if span := trace.FromContext(ctx); span != nil {
				if s, ok := status.FromError(err); ok {
					span.SetStatus(trace.Status{Code: int32(s.Code()), Message: s.Message()})
				} else {
					span.SetStatus(trace.Status{Code: int32(codes.Unknown), Message: err.Error()})
				}
				span.End()
			}
		},
	)

	return func(c *kitgrpc.Client) {
		clientBefore(c)
		clientFinalizer(c)
	}
}

// GRPCServerTrace enables OpenCensus tracing of a Go kit gRPC transport server.
func GRPCServerTrace(options ...TracerOption) kitgrpc.ServerOption {
	cfg := TracerOptions{
		sampler: trace.AlwaysSample(),
	}

	for _, option := range options {
		option(&cfg)
	}

	serverBefore := kitgrpc.ServerBefore(
		func(ctx context.Context, md metadata.MD) context.Context {
			var name string

			if cfg.name != "" {
				name = cfg.name
			} else {
				name, _ = ctx.Value(kitgrpc.ContextKeyRequestMethod).(string)
				if name == "" {
					// we can't find the gRPC method. probably the
					// unaryInterceptor was not wired up.
					name = "unknown grpc method"
				}
			}

			var (
				parentContext trace.SpanContext
				traceContext  = md[propagationKey]
				ok            bool
			)

			if len(traceContext) > 0 {
				traceContextBinary := []byte(traceContext[0])
				parentContext, ok = propagation.FromBinary(traceContextBinary)
				if ok && !cfg.public {
					ctx, _ = trace.StartSpanWithRemoteParent(
						ctx,
						name,
						parentContext,
						trace.WithSpanKind(trace.SpanKindServer),
						trace.WithSampler(cfg.sampler),
					)
					return ctx
				}
			}
			ctx, span := trace.StartSpan(
				ctx,
				name,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithSampler(cfg.sampler),
			)
			if ok {
				span.AddLink(
					trace.Link{
						TraceID: parentContext.TraceID,
						SpanID:  parentContext.SpanID,
						Type:    trace.LinkTypeChild,
					},
				)
			}
			return ctx
		},
	)

	serverFinalizer := kitgrpc.ServerFinalizer(
		func(ctx context.Context, err error) {
			if span := trace.FromContext(ctx); span != nil {
				if s, ok := status.FromError(err); ok {
					span.SetStatus(trace.Status{Code: int32(s.Code()), Message: s.Message()})
				} else {
					span.SetStatus(trace.Status{Code: int32(codes.Internal), Message: err.Error()})
				}
				span.End()
			}
		},
	)

	return func(s *kitgrpc.Server) {
		serverBefore(s)
		serverFinalizer(s)
	}
}
