package test

import (
	"context"
	"errors"

	pb "github.com/go-kit/kit/transport/grpc/_pb"
)

func encodeRequest(ctx context.Context, req interface{}) (interface{}, error) {
	r, ok := req.(TestRequest)
	if !ok {
		return nil, errors.New("request encode error")
	}
	return &pb.TestRequest{A: r.A, B: r.B}, nil
}

func decodeRequest(ctx context.Context, req interface{}) (interface{}, error) {
	r, ok := req.(*pb.TestRequest)
	if !ok {
		return nil, errors.New("request decode error")
	}
	return TestRequest{A: r.A, B: r.B}, nil
}

func encodeResponse(ctx context.Context, resp interface{}) (interface{}, error) {
	r, ok := resp.(*TestResponse)
	if !ok {
		return nil, errors.New("response encode error")
	}
	return &pb.TestResponse{V: r.V}, nil
}

func decodeResponse(ctx context.Context, resp interface{}) (interface{}, error) {
	r, ok := resp.(*pb.TestResponse)
	if !ok {
		return nil, errors.New("response decode error")
	}
	return &TestResponse{V: r.V, Ctx: ctx}, nil
}
