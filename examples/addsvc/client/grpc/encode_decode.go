package grpc

import (
	"golang.org/x/net/context"

	"github.com/go-kit/kit/examples/addsvc/pb"
	"github.com/go-kit/kit/examples/addsvc/server"
)

func encodeSumRequest(ctx context.Context, request interface{}) (interface{}, error) {
	req := request.(server.SumRequest)
	return &pb.SumRequest{
		A: int64(req.A),
		B: int64(req.B),
	}, nil
}

func encodeConcatRequest(ctx context.Context, request interface{}) (interface{}, error) {
	req := request.(server.ConcatRequest)
	return &pb.ConcatRequest{
		A: req.A,
		B: req.B,
	}, nil
}

func decodeSumResponse(ctx context.Context, response interface{}) (interface{}, error) {
	resp := response.(*pb.SumReply)
	return server.SumResponse{
		V: int(resp.V),
	}, nil
}

func decodeConcatResponse(ctx context.Context, response interface{}) (interface{}, error) {
	resp := response.(*pb.ConcatReply)
	return server.ConcatResponse{
		V: resp.V,
	}, nil
}
