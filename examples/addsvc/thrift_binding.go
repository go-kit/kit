package main

import (
	"github.com/go-kit/kit/examples/addsvc/server"
	thriftadd "github.com/go-kit/kit/examples/addsvc/thrift/gen-go/add"
)

type thriftBinding struct {
	server.AddService
}

func (tb thriftBinding) Sum(a, b int64) (*thriftadd.SumReply, error) {
	v := tb.AddService.Sum(int(a), int(b))
	return &thriftadd.SumReply{Value: int64(v)}, nil
}

func (tb thriftBinding) Concat(a, b string) (*thriftadd.ConcatReply, error) {
	v := tb.AddService.Concat(a, b)
	return &thriftadd.ConcatReply{Value: v}, nil
}
