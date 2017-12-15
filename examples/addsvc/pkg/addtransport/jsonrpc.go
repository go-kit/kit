package addtransport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/examples/addsvc/pkg/addendpoint"
	"github.com/go-kit/kit/examples/addsvc/pkg/addservice"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport/http/jsonrpc"
)

// NewJSONRPCHandler returns a JSON RPC Server/Handler that can be passed to http.Handle()
func NewJSONRPCHandler(endpoints addendpoint.Set, logger log.Logger) *jsonrpc.Server {
	handler := jsonrpc.NewServer(
		makeEndpointCodecMap(endpoints),
		jsonrpc.ServerErrorLogger(logger),
	)
	return handler
}

// NewJSONRPCClient returns an addservice backed by a JSON RPC over HTTP server
// living at the remote instance. We expect instance to come from a service
// discovery system, so likely of the form "host:port". We bake-in certain
// middlewares, implementing the client library pattern.
func NewJSONRPCClient(instance string, logger log.Logger) (addservice.Service, error) {
	tgt, err := url.Parse(instance)
	if err != nil {
		return nil, err
	}

	var sumEndpoint endpoint.Endpoint
	{
		c := jsonrpc.NewClient(tgt, "sum", encodeSumRequest, decodeSumResponse)
		sumEndpoint = c.Endpoint()
		// TODO: Add middlewares.
	}

	var concatEndpoint endpoint.Endpoint
	{
	}

	// Returning the endpoint.Set as a service.Service relies on the
	// endpoint.Set implementing the Service methods. That's just a simple bit
	// of glue code.
	return addendpoint.Set{
		SumEndpoint:    sumEndpoint,
		ConcatEndpoint: concatEndpoint,
	}, nil

}

// makeEndpointCodecMap returns a codec map configured for the addsvc.
func makeEndpointCodecMap(endpoints addendpoint.Set) jsonrpc.EndpointCodecMap {
	return jsonrpc.EndpointCodecMap{
		"sum": jsonrpc.EndpointCodec{
			Endpoint: endpoints.SumEndpoint,
			Decode:   decodeSumRequest,
			Encode:   encodeSumResponse,
		},
	}
}

func decodeSumRequest(_ context.Context, msg json.RawMessage) (interface{}, error) {
	var req addendpoint.SumRequest
	err := json.Unmarshal(msg, &req)
	if err != nil {
		return nil, &jsonrpc.Error{
			Code:    -32000,
			Message: fmt.Sprintf("couldn't unmarshal body to sum request: %s", err),
		}
	}
	return req, nil
}

func encodeSumResponse(_ context.Context, obj interface{}) (json.RawMessage, error) {
	res, ok := obj.(addendpoint.SumResponse)
	if !ok {
		return nil, &jsonrpc.Error{
			Code:    -32000,
			Message: fmt.Sprintf("Asserting result to *SumResponse failed. Got %T, %+v", obj, obj),
		}
	}
	b, err := json.Marshal(res)
	if err != nil {
		return nil, fmt.Errorf("couldn't marshal response: %s", err)
	}
	return b, nil
}

func decodeSumResponse(_ context.Context, msg json.RawMessage) (interface{}, error) {
	var res addendpoint.SumResponse
	err := json.Unmarshal(msg, &res)
	if err != nil {
		return nil, fmt.Errorf("couldn't unmarshal body to SumResponse: %s", err)
	}
	return res, nil
}

func encodeSumRequest(_ context.Context, obj interface{}) (json.RawMessage, error) {
	req, ok := obj.(addendpoint.SumRequest)
	if !ok {
		return nil, fmt.Errorf("couldn't assert request as SumRequest, got %T", obj)
	}
	b, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("couldn't marshal request: %s", err)
	}
	return b, nil
}
