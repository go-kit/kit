package awslambda

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

// Server wraps an endpoint.
type Server struct {
	e            endpoint.Endpoint
	dec          DecodeRequestFunc
	enc          EncodeResponseFunc
	before       []ServerRequestFunc
	after        []ServerResponseFunc
	errorEncoder ErrorEncoder
	logger       log.Logger
}

// NewServer constructs a new server, which implements the rules
// of handling AWS APIGatewayProxy event with AWS lambda.
func NewServer(
	e endpoint.Endpoint,
	dec DecodeRequestFunc,
	enc EncodeResponseFunc,
	options ...ServerOption,
) *Server {
	s := &Server{
		e:      e,
		dec:    dec,
		enc:    enc,
		logger: log.NewNopLogger(),
	}
	for _, option := range options {
		option(s)
	}
	return s
}

// ServerOption sets an optional parameter for servers.
type ServerOption func(*Server)

// ServerBefore functions are executed on the AWS APIGatewayProxy
// request object before the request is decoded.
func ServerBefore(before ...ServerRequestFunc) ServerOption {
	return func(s *Server) { s.before = append(s.before, before...) }
}

// ServerAfter functions are only executed after invoking the endpoint
// but prior to returning a response.
func ServerAfter(after ...ServerResponseFunc) ServerOption {
	return func(s *Server) { s.after = append(s.after, after...) }
}

// ServerErrorLogger is used to log non-terminal errors.
// By default, no errors are logged.
func ServerErrorLogger(logger log.Logger) ServerOption {
	return func(s *Server) { s.logger = logger }
}

// ServerErrorEncoder is used to encode errors for
// AWS APIGatewayProxy response.
func ServerErrorEncoder(ee ErrorEncoder) ServerOption {
	return func(s *Server) { s.errorEncoder = ee }
}

// ServeHTTPLambda implements the rules of AWS lambda handler of
// AWS APIGatewayProxy event.
func (s *Server) ServeHTTPLambda(
	ctx context.Context, req events.APIGatewayProxyRequest,
) (
	resp events.APIGatewayProxyResponse, err error,
) {
	for _, f := range s.before {
		ctx = f(ctx, req)
	}

	request, err := s.dec(ctx, req)
	if err != nil {
		s.logger.Log("err", err)
		resp, err = s.errorEncoder(ctx, err, resp)
		return
	}

	response, err := s.e(ctx, request)
	if err != nil {
		s.logger.Log("err", err)
		resp, err = s.errorEncoder(ctx, err, resp)
		return
	}

	for _, f := range s.after {
		ctx = f(ctx, resp)
	}

	if resp, err = s.enc(ctx, response, resp); err != nil {
		s.logger.Log("err", err)
		resp, err = s.errorEncoder(ctx, err, resp)
		return
	}

	return resp, err
}
