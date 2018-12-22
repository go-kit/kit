package awslambda

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

// Server wraps an endpoint.
type Server struct {
	e      endpoint.Endpoint
	dec    DecodeRequestFunc
	enc    EncodeResponseFunc
	logger log.Logger
}

// NewServer constructs a new server, which implements the rules
// of handling AWS API gateway event with AWS lambda.
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

// ServerErrorLogger is used to log non-terminal errors. By default, no errors
// are logged.
func ServerErrorLogger(logger log.Logger) ServerOption {
	return func(s *Server) { s.logger = logger }
}

// ServeHTTPLambda implements the rules of AWS lambda handler of
// AWS API gateway event.
func (s *Server) ServeHTTPLambda(
	ctx context.Context, req events.APIGatewayProxyRequest,
) (
	resp events.APIGatewayProxyResponse, err error,
) {
	request, err := s.dec(ctx, req)
	if err != nil {
		s.logger.Log("err", err)
		return
	}

	response, err := s.e(ctx, request)
	if err != nil {
		s.logger.Log("err", err)
		return
	}

	if resp, err = s.enc(ctx, response); err != nil {
		s.logger.Log("err", err)
		return
	}

	return resp, err
}
