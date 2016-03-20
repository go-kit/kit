package main

import (
	"golang.org/x/net/context"

	"github.com/go-kit/kit/examples/addsvc/pb"
	"github.com/go-kit/kit/examples/addsvc/server"
	servergrpc "github.com/go-kit/kit/examples/addsvc/server/grpc"
	"github.com/go-kit/kit/transport/grpc"
)

type grpcBinding struct {
	sum, concat grpc.Handler
}

func newGRPCBinding(ctx context.Context, svc server.AddService) grpcBinding {
	return grpcBinding{
		sum:    grpc.NewServer(ctx, makeSumEndpoint(svc), servergrpc.DecodeSumRequest, servergrpc.EncodeSumResponse),
		concat: grpc.NewServer(ctx, makeConcatEndpoint(svc), servergrpc.DecodeConcatRequest, servergrpc.EncodeConcatResponse),
	}
}

func (b grpcBinding) Sum(ctx context.Context, req *pb.SumRequest) (*pb.SumReply, error) {
	_, resp, err := b.sum.ServeGRPC(ctx, req)
	return resp.(*pb.SumReply), err
}

func (b grpcBinding) Concat(ctx context.Context, req *pb.ConcatRequest) (*pb.ConcatReply, error) {
	_, resp, err := b.concat.ServeGRPC(ctx, req)
	return resp.(*pb.ConcatReply), err
}
