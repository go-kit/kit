package grpc

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/examples/addsvc/pb"
	"github.com/go-kit/kit/examples/addsvc/server"
	"github.com/go-kit/kit/log"
	grpctransport "github.com/go-kit/kit/transport/grpc"
)

// New returns an AddService that's backed by the provided ClientConn.
func New(ctx context.Context, cc *grpc.ClientConn, logger log.Logger) server.AddService {
	return client{
		Context: ctx,
		Logger:  logger,
		sum:     grpctransport.NewClient(cc, "Add", "sum", encodeSumRequest, decodeSumResponse, pb.SumReply{}).Endpoint(),
		concat:  grpctransport.NewClient(cc, "Add", "concat", encodeConcatRequest, decodeConcatResponse, pb.ConcatReply{}).Endpoint(),
	}
}

type client struct {
	context.Context
	log.Logger
	sum    endpoint.Endpoint
	concat endpoint.Endpoint
}

// TODO(pb): If your service interface methods don't return an error, we have
// no way to signal problems with a service client. If they don't take a
// context, we have to provide a global context for any transport that
// requires one, effectively making your service a black box to any context-
// specific information. So, we should make some recommendations:
//
// - To get started, a simple service interface is probably fine.
//
// - To properly deal with transport errors, every method on your service
//   should return an error. This is probably important.
//
// - To properly deal with context information, every method on your service
//   can take a context as its first argument. This may or may not be
//   important.

func (c client) Sum(a, b int) int {
	request := &server.SumRequest{
		A: a,
		B: b,
	}
	reply, err := c.sum(c.Context, request)
	if err != nil {
		c.Logger.Log("err", err) // Without an error return parameter, we can't do anything else...
		return 0
	}

	r := reply.(server.SumResponse)
	return r.V
}

func (c client) Concat(a, b string) string {
	request := &server.ConcatRequest{
		A: a,
		B: b,
	}
	reply, err := c.concat(c.Context, request)
	if err != nil {
		c.Logger.Log("err", err) // Without an error return parameter, we can't do anything else...
		return ""
	}

	r := reply.(server.ConcatResponse)
	return r.V
}
