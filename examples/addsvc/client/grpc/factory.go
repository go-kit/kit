package grpc

import (
	"io"

	kitot "github.com/go-kit/kit/tracing/opentracing"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/examples/addsvc/pb"
	"github.com/go-kit/kit/loadbalancer"
	"github.com/go-kit/kit/log"
	grpctransport "github.com/go-kit/kit/transport/grpc"
)

// MakeSumEndpointFactory returns a loadbalancer.Factory that transforms GRPC
// host:port strings into Endpoints that call the Sum method on a GRPC server
// at that address.
func MakeSumEndpointFactory(tracer opentracing.Tracer, tracingLogger log.Logger) loadbalancer.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		cc, err := grpc.Dial(instance, grpc.WithInsecure())
		return grpctransport.NewClient(
			cc,
			"Add",
			"Sum",
			encodeSumRequest,
			decodeSumResponse,
			pb.SumReply{},
			grpctransport.SetClientBefore(kitot.ToGRPCRequest(tracer, tracingLogger)),
		).Endpoint(), cc, err
	}
}

// MakeConcatEndpointFactory returns a loadbalancer.Factory that transforms
// GRPC host:port strings into Endpoints that call the Concat method on a GRPC
// server at that address.
func MakeConcatEndpointFactory(tracer opentracing.Tracer, tracingLogger log.Logger) loadbalancer.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		cc, err := grpc.Dial(instance, grpc.WithInsecure())
		return grpctransport.NewClient(
			cc,
			"Add",
			"Concat",
			encodeConcatRequest,
			decodeConcatResponse,
			pb.ConcatReply{},
			grpctransport.SetClientBefore(kitot.ToGRPCRequest(tracer, tracingLogger)),
		).Endpoint(), cc, err
	}
}
