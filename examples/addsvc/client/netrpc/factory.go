package netrpc

import (
	"io"
	"net/rpc"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/examples/addsvc/server"
)

// SumEndpointFactory transforms host:port strings into Endpoints.
func SumEndpointFactory(instance string) (endpoint.Endpoint, io.Closer, error) {
	client, err := rpc.DialHTTP("tcp", instance)
	if err != nil {
		return nil, nil, err
	}

	return func(ctx context.Context, request interface{}) (interface{}, error) {
		var reply server.SumResponse
		if err := client.Call("addsvc.Sum", request.(server.SumRequest), &reply); err != nil {
			return server.SumResponse{}, err
		}
		return reply, nil
	}, client, nil
}

// ConcatEndpointFactory transforms host:port strings into Endpoints.
func ConcatEndpointFactory(instance string) (endpoint.Endpoint, io.Closer, error) {
	client, err := rpc.DialHTTP("tcp", instance)
	if err != nil {
		return nil, nil, err
	}

	return func(ctx context.Context, request interface{}) (interface{}, error) {
		var reply server.ConcatResponse
		if err := client.Call("addsvc.Concat", request.(server.ConcatRequest), &reply); err != nil {
			return server.ConcatResponse{}, err
		}
		return reply, nil
	}, client, nil
}
