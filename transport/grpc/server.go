package grpc

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

// Handler which should be called from the grpc binding of the service
// implementation.
type Handler interface {
	ServeGRPC(context.Context, interface{}) (context.Context, interface{}, error)
}

// Server wraps an endpoint and implements grpc.Handler.
type Server struct {
	ctx    context.Context
	e      endpoint.Endpoint
	dec    DecodeRequestFunc
	enc    EncodeResponseFunc
	before []RequestFunc
	after  []ResponseFunc
	logger log.Logger
}

// NewServer constructs a new server, which implements grpc.Server and wraps
// the provided endpoint.
func NewServer(
	ctx context.Context,
	e endpoint.Endpoint,
	dec DecodeRequestFunc,
	enc EncodeResponseFunc,
	options ...ServerOption,
) *Server {
	s := &Server{
		ctx:    ctx,
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
func ServerBefore(before ...RequestFunc) ServerOption {
	return func(s *Server) { s.before = before }
}

// ServerAfter functions are executed on the HTTP response writer after the
// endpoint is invoked, but before anything is written to the client.
func ServerAfter(after ...ResponseFunc) ServerOption {
	return func(s *Server) { s.after = after }
}

// ServerErrorLogger is used to log non-terminal errors. By default, no errors
// are logged.
func ServerErrorLogger(logger log.Logger) ServerOption {
	return func(s *Server) { s.logger = logger }
}

// ServeGRPC implements grpc.Handler
func (s Server) ServeGRPC(grpcCtx context.Context, r interface{}) (context.Context, interface{}, error) {
	ctx, cancel := context.WithCancel(s.ctx)
	defer cancel()

	// retrieve gRPC metadata
	md, ok := metadata.FromContext(grpcCtx)
	if !ok {
		md = metadata.MD{}
	}

	for _, f := range s.before {
		ctx = f(ctx, &md)
	}

	// store potentially updated metadata in the gRPC context
	grpcCtx = metadata.NewContext(grpcCtx, md)

	request, err := s.dec(grpcCtx, r)
	if err != nil {
		s.logger.Log("err", err)
		return grpcCtx, nil, BadRequestError{err}
	}

	response, err := s.e(ctx, request)
	if err != nil {
		s.logger.Log("err", err)
		return grpcCtx, nil, err
	}

	for _, f := range s.after {
		f(ctx, &md)
	}

	// store potentially updated metadata in the gRPC context
	grpcCtx = metadata.NewContext(grpcCtx, md)

	grpcResp, err := s.enc(grpcCtx, response)
	if err != nil {
		s.logger.Log("err", err)
		return grpcCtx, nil, err
	}
	return grpcCtx, grpcResp, nil
}

// BadRequestError is an error in decoding the request.
type BadRequestError struct {
	Err error
}

// Error implements the error interface.
func (err BadRequestError) Error() string {
	return err.Err.Error()
}
