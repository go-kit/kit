package awslambda

import (
	"context"

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
	finalizer    []ServerFinalizerFunc
	logger       log.Logger
}

// NewServer constructs a new server, which implements
// the AWS lambda.Handler interface.
func NewServer(
	e endpoint.Endpoint,
	dec DecodeRequestFunc,
	enc EncodeResponseFunc,
	options ...ServerOption,
) *Server {
	s := &Server{
		e:            e,
		dec:          dec,
		enc:          enc,
		logger:       log.NewNopLogger(),
		errorEncoder: DefaultErrorEncoder,
	}
	for _, option := range options {
		option(s)
	}
	return s
}

// ServerOption sets an optional parameter for servers.
type ServerOption func(*Server)

// ServerBefore functions are executed on the payload byte,
// before the request is decoded.
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

// ServerErrorEncoder is used to encode errors.
func ServerErrorEncoder(ee ErrorEncoder) ServerOption {
	return func(s *Server) { s.errorEncoder = ee }
}

// ServerFinalizer sets finalizer which are called at the end of
// request. By default no finalizer is registered.
func ServerFinalizer(f ...ServerFinalizerFunc) ServerOption {
	return func(s *Server) { s.finalizer = append(s.finalizer, f...) }
}

// DefaultErrorEncoder defines the default behavior of encoding an error response,
// where it returns nil, and the error itself.
func DefaultErrorEncoder(ctx context.Context, err error) ([]byte, error) {
	return nil, err
}

// Invoke represents implementation of the AWS lambda.Handler interface.
func (s *Server) Invoke(
	ctx context.Context,
	payload []byte,
) (resp []byte, err error) {
	if len(s.finalizer) > 0 {
		defer func() {
			for _, f := range s.finalizer {
				f(ctx, resp, err)
			}
		}()
	}

	for _, f := range s.before {
		ctx = f(ctx, payload)
	}

	request, err := s.dec(ctx, payload)
	if err != nil {
		s.logger.Log("err", err)
		resp, err = s.errorEncoder(ctx, err)
		return
	}

	response, err := s.e(ctx, request)
	if err != nil {
		s.logger.Log("err", err)
		resp, err = s.errorEncoder(ctx, err)
		return
	}

	for _, f := range s.after {
		ctx = f(ctx, response)
	}

	if resp, err = s.enc(ctx, response); err != nil {
		s.logger.Log("err", err)
		resp, err = s.errorEncoder(ctx, err)
		return
	}

	return resp, err
}
