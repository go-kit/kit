package grpc

import (
	"github.com/go-kit/kit/examples/addsvc/pb"
	"github.com/go-kit/kit/examples/addsvc/server"
	"golang.org/x/net/context"
)

func encodeSumRequest(ctx context.Context, req interface{}) (interface{}, error) {
	sumRequest := req.(server.SumRequest)

	pbRequest := &pb.SumRequest{
		A: int64(sumRequest.A),
		B: int64(sumRequest.B),
	}
	return pbRequest, nil
}

func encodeConcatRequest(ctx context.Context, req interface{}) (interface{}, error) {
	concatRequest := req.(server.ConcatRequest)

	pbRequest := &pb.ConcatRequest{
		A: concatRequest.A,
		B: concatRequest.B,
	}
	return pbRequest, nil
}

func decodeSumResponse(ctx context.Context, resp interface{}) (interface{}, error) {
	pbResponse := resp.(pb.SumReply)

	sumResponse := &server.SumResponse{
		V: int(pbResponse.V),
	}
	return sumResponse, nil
}

func decodeConcatResponse(ctx context.Context, resp interface{}) (interface{}, error) {
	pbResponse := resp.(pb.ConcatReply)

	concatResponse := &server.ConcatResponse{
		V: pbResponse.V,
	}
	return concatResponse, nil
}
