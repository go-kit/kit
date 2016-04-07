package grpc

import (
	"io"

	"google.golang.org/grpc"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/examples/addsvc/pb"
	grpctransport "github.com/go-kit/kit/transport/grpc"
)

// SumEndpointFactory transforms GRPC host:port strings into Endpoints that call the Sum method on a GRPC server
// at that address.
func SumEndpointFactory(instance string) (endpoint.Endpoint, io.Closer, error) {
	cc, err := grpc.Dial(instance, grpc.WithInsecure())
	return grpctransport.NewClient(
		cc,
		"Add",
		"Sum",
		encodeSumRequest,
		decodeSumResponse,
		pb.SumReply{},
	).Endpoint(), cc, err
}

// ConcatEndpointFactory transforms GRPC host:port strings into Endpoints that call the Concat method on a GRPC server
// at that address.
func ConcatEndpointFactory(instance string) (endpoint.Endpoint, io.Closer, error) {
	cc, err := grpc.Dial(instance, grpc.WithInsecure())
	return grpctransport.NewClient(
		cc,
		"Add",
		"Concat",
		encodeConcatRequest,
		decodeConcatResponse,
		pb.ConcatReply{},
	).Endpoint(), cc, err
}
