package transport

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	addendpoint "github.com/go-kit/kit/examples/addsvc2/pkg/endpoint"
	"github.com/go-kit/kit/examples/addsvc2/pkg/service"
	thriftadd "github.com/go-kit/kit/examples/addsvc2/thrift/gen-go/addsvc"
)

// MakeThriftHandler makes a set of endpoints available as a Thrift service.
func MakeThriftHandler(ctx context.Context, endpoints addendpoint.Set) thriftadd.AddService {
	return &thriftServer{
		ctx:       ctx,
		endpoints: endpoints,
	}
}

type thriftServer struct {
	ctx       context.Context
	endpoints addendpoint.Set
}

func (s *thriftServer) Sum(a int64, b int64) (*thriftadd.SumReply, error) {
	request := addendpoint.SumRequest{A: int(a), B: int(b)}
	response, err := s.endpoints.SumEndpoint(s.ctx, request)
	if err != nil {
		return nil, err
	}
	resp := response.(addendpoint.SumResponse)
	return &thriftadd.SumReply{Value: int64(resp.V), Err: err2str(resp.Err)}, nil
}

func (s *thriftServer) Concat(a string, b string) (*thriftadd.ConcatReply, error) {
	request := addendpoint.ConcatRequest{A: a, B: b}
	response, err := s.endpoints.ConcatEndpoint(s.ctx, request)
	if err != nil {
		return nil, err
	}
	resp := response.(addendpoint.ConcatResponse)
	return &thriftadd.ConcatReply{Value: resp.V, Err: err2str(resp.Err)}, nil
}

// MakeThriftSumEndpoint returns an endpoint that invokes the passed Thrift client.
// Useful only in clients, and only until a proper transport/thrift.Client exists.
func MakeThriftSumEndpoint(client *thriftadd.AddServiceClient) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(addendpoint.SumRequest)
		reply, err := client.Sum(int64(req.A), int64(req.B))
		if err == service.ErrIntOverflow {
			return nil, err // special case; see comment on ErrIntOverflow
		}
		return addendpoint.SumResponse{V: int(reply.Value), Err: err}, nil
	}
}

// MakeThriftConcatEndpoint returns an endpoint that invokes the passed Thrift
// client. Useful only in clients, and only until a proper
// transport/thrift.Client exists.
func MakeThriftConcatEndpoint(client *thriftadd.AddServiceClient) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(addendpoint.ConcatRequest)
		reply, err := client.Concat(req.A, req.B)
		return addendpoint.ConcatResponse{V: reply.Value, Err: err}, nil
	}
}
