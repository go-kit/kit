package twirp

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

// Handler which should be called from the Twirp binding of the service
// implementation. The incoming request parameter, and returned response
// parameter, are both Twirp types, not user-domain.
type Handler interface {
	ServeTwirp(ctx context.Context, request interface{}) (context.Context, interface{}, error)
}

// Server wraps an endpoint and implements Twirp Handler.
type Server struct {
	e      endpoint.Endpoint
	dec    DecodeRequestFunc
	enc    EncodeResponseFunc
	before []ServerRequestFunc
	after  []ServerResponseFunc
	logger log.Logger
}

// NewServer constructs a new server, which implements wraps the provided
// endpoint and implements the Handler interface. Consumers should write
// bindings that adapt the concrete Twirp methods from their compiled protobuf
// definitions to individual handlers. Request and response objects are from the
// caller business domain, not Twirp request and reply types.
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

// ServerBefore functions are executed on the HTTP request object before the
// request is decoded.
func ServerBefore(before ...ServerRequestFunc) ServerOption {
	return func(s *Server) { s.before = append(s.before, before...) }
}

// ServerAfter functions are executed on the HTTP response writer after the
// endpoint is invoked, but before anything is written to the client.
func ServerAfter(after ...ServerResponseFunc) ServerOption {
	return func(s *Server) { s.after = append(s.after, after...) }
}

// ServerErrorLogger is used to log non-terminal errors. By default, no errors
// are logged.
func ServerErrorLogger(logger log.Logger) ServerOption {
	return func(s *Server) { s.logger = logger }
}

// ServeTwirp implements the Handler interface.
func (s Server) ServeTwirp(ctx context.Context, req interface{}) (context.Context, interface{}, error) {

	// Process ServerRequestFunctions
	for _, f := range s.before {
		ctx = f(ctx)
	}
	request, err := s.dec(ctx, req)
	if err != nil {
		s.logger.Log("err", err)
		return ctx, nil, err
	}

	response, err := s.e(ctx, request)
	if err != nil {
		s.logger.Log("err", err)
		return ctx, nil, err
	}

	// Process ServerResponseFunctions
	for _, f := range s.after {
		ctx, err = f(ctx)
		if err != nil {
			return ctx, nil, err
		}
	}

	twirpResp, err := s.enc(ctx, response)
	if err != nil {
		s.logger.Log("err", err)
		return ctx, nil, err
	}

	return ctx, twirpResp, nil
}
