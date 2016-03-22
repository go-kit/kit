package grpc

import (
	"golang.org/x/net/context"

	"github.com/go-kit/kit/examples/addsvc/pb"
	"github.com/go-kit/kit/examples/addsvc/server"
)

func DecodeSumRequest(ctx context.Context, req interface{}) (interface{}, error) {
	sumRequest := req.(*pb.SumRequest)

	return &server.SumRequest{
		A: int(sumRequest.A),
		B: int(sumRequest.B),
	}, nil
}

func DecodeConcatRequest(ctx context.Context, req interface{}) (interface{}, error) {
	concatRequest := req.(*pb.ConcatRequest)

	return &server.ConcatRequest{
		A: concatRequest.A,
		B: concatRequest.B,
	}, nil
}

func EncodeSumResponse(ctx context.Context, resp interface{}) (interface{}, error) {
	domainResponse := resp.(server.SumResponse)

	return &pb.SumReply{
		V: int64(domainResponse.V),
	}, nil
}

func EncodeConcatResponse(ctx context.Context, resp interface{}) (interface{}, error) {
	domainResponse := resp.(server.ConcatResponse)

	return &pb.ConcatReply{
		V: domainResponse.V,
	}, nil
}
