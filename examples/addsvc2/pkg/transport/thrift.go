package transport

import (
	"context"
	"time"

	jujuratelimit "github.com/juju/ratelimit"
	"github.com/sony/gobreaker"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/ratelimit"

	addendpoint "github.com/go-kit/kit/examples/addsvc2/pkg/endpoint"
	addservice "github.com/go-kit/kit/examples/addsvc2/pkg/service"
	thriftadd "github.com/go-kit/kit/examples/addsvc2/thrift/gen-go/addsvc"
)

type thriftServer struct {
	ctx       context.Context
	endpoints addendpoint.Set
}

// NewThriftServer makes a set of endpoints available as a Thrift service.
func NewThriftServer(ctx context.Context, endpoints addendpoint.Set) thriftadd.AddService {
	return &thriftServer{
		ctx:       ctx,
		endpoints: endpoints,
	}
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

// NewThriftClient returns an AddService backed by a Thrift server described by
// the provided client. The caller is responsible for constructing the client,
// and eventually closing the underlying transport.
func NewThriftClient(client *thriftadd.AddServiceClient) addservice.Service {
	// We construct a single ratelimiter middleware, to limit the total outgoing
	// QPS from this client to all methods on the remote instance. We also
	// construct per-endpoint circuitbreaker middlewares to demonstrate how
	// that's done, although they could easily be combined into a single breaker
	// for the entire remote instance, too.
	limiter := ratelimit.NewTokenBucketLimiter(jujuratelimit.NewBucketWithRate(100, 100))

	// Each individual endpoint is an http/transport.Client (which implements
	// endpoint.Endpoint) that gets wrapped with various middlewares. If you
	// could rely on a consistent set of client behavior.
	var sumEndpoint endpoint.Endpoint
	{
		sumEndpoint = MakeThriftSumEndpoint(client)
		sumEndpoint = limiter(sumEndpoint)
		sumEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "Sum",
			Timeout: 30 * time.Second,
		}))(sumEndpoint)
	}

	// The Concat endpoint is the same thing, with slightly different
	// middlewares to demonstrate how to specialize per-endpoint.
	var concatEndpoint endpoint.Endpoint
	{
		concatEndpoint = MakeThriftConcatEndpoint(client)
		concatEndpoint = limiter(concatEndpoint)
		concatEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "Concat",
			Timeout: 10 * time.Second,
		}))(concatEndpoint)
	}

	// Returning the endpoint.Set as a service.Service relies on the
	// endpoint.Set implementing the Service methods. That's just a simple bit
	// of glue code.
	return addendpoint.Set{
		SumEndpoint:    sumEndpoint,
		ConcatEndpoint: concatEndpoint,
	}
}

// MakeThriftSumEndpoint returns an endpoint that invokes the passed Thrift client.
// Useful only in clients, and only until a proper transport/thrift.Client exists.
func MakeThriftSumEndpoint(client *thriftadd.AddServiceClient) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(addendpoint.SumRequest)
		reply, err := client.Sum(int64(req.A), int64(req.B))
		if err == addservice.ErrIntOverflow {
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
