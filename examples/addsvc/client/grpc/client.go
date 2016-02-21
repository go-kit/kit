package grpc

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/go-kit/kit/examples/addsvc/pb"
	"github.com/go-kit/kit/examples/addsvc/server"
	"github.com/go-kit/kit/log"
)

// New returns an AddService that's backed by the provided ClientConn.
func New(ctx context.Context, cc *grpc.ClientConn, logger log.Logger) server.AddService {
	return client{ctx, pb.NewAddClient(cc), logger}
}

type client struct {
	context.Context
	pb.AddClient
	log.Logger
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
	request := &pb.SumRequest{
		A: int64(a),
		B: int64(b),
	}
	reply, err := c.AddClient.Sum(c.Context, request)
	if err != nil {
		c.Logger.Log("err", err) // Without an error return parameter, we can't do anything else...
		return 0
	}
	return int(reply.V)
}

func (c client) Concat(a, b string) string {
	request := &pb.ConcatRequest{
		A: a,
		B: b,
	}
	reply, err := c.AddClient.Concat(c.Context, request)
	if err != nil {
		c.Logger.Log("err", err)
		return ""
	}
	return reply.V
}
