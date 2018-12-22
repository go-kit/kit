package awslambda

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/go-kit/kit/endpoint"
)

// Server wraps an endpoint.
type Server struct {
	e   endpoint.Endpoint
	dec DecodeRequestFunc
	enc EncodeResponseFunc
}

// NewServer constructs a new server, which implements the rules
// of handling AWS API gateway event with AWS lambda.
func NewServer(
	e endpoint.Endpoint,
	dec DecodeRequestFunc,
	enc EncodeResponseFunc,
) *Server {
	s := &Server{
		e:   e,
		dec: dec,
		enc: enc,
	}
	return s
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
		return
	}

	response, err := s.e(ctx, request)
	if err != nil {
		return
	}

	if resp, err = s.enc(ctx, response); err != nil {
		return
	}

	return resp, err
}
