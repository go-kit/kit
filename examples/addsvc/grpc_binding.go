package main

import (
	"golang.org/x/net/context"

	"github.com/go-kit/kit/examples/addsvc/pb"
	"github.com/go-kit/kit/examples/addsvc/server"
)

type grpcBinding struct {
	server.AddService
}

func (b grpcBinding) Sum(ctx context.Context, req *pb.SumRequest) (*pb.SumReply, error) {
	return &pb.SumReply{V: int64(b.AddService.Sum(int(req.A), int(req.B)))}, nil
}

func (b grpcBinding) Concat(ctx context.Context, req *pb.ConcatRequest) (*pb.ConcatReply, error) {
	return &pb.ConcatReply{V: b.AddService.Concat(req.A, req.B)}, nil
}
