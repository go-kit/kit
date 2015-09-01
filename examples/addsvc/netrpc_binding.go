package main

import (
	"github.com/go-kit/kit/examples/addsvc/server"
)

type netrpcBinding struct {
	server.AddService
}

func (b netrpcBinding) Sum(request server.SumRequest, response *server.SumResponse) error {
	v := b.AddService.Sum(request.A, request.B)
	(*response) = server.SumResponse{V: v}
	return nil
}

func (b netrpcBinding) Concat(request server.ConcatRequest, response *server.ConcatResponse) error {
	v := b.AddService.Concat(request.A, request.B)
	(*response) = server.ConcatResponse{V: v}
	return nil
}
