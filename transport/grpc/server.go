package grpc

import (
	oldcontext "golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

// Handler which should be called from the gRPC binding of the service
// implementation. The incoming request parameter, and returned response
// parameter, are both gRPC types, not user-domain.
type Handler interface {
	ServeGRPC(ctx oldcontext.Context, request interface{}) (oldcontext.Context, interface{}, error)
}

// Server wraps an endpoint and implements grpc.Handler.
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
// bindings that adapt the concrete gRPC methods from their compiled protobuf
// definitions to individual handlers. Request and response objects are from the
// caller business domain, not gRPC request and reply types.
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

// ServeGRPC implements the Handler interface.
func (s Server) ServeGRPC(ctx oldcontext.Context, req interface{}) (oldcontext.Context, interface{}, error) {
	// Retrieve gRPC metadata.
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}

	for _, f := range s.before {
		ctx = f(ctx, md)
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

	var mdHeader, mdTrailer metadata.MD
	for _, f := range s.after {
		ctx = f(ctx, &mdHeader, &mdTrailer)
	}

	grpcResp, err := s.enc(ctx, response)
	if err != nil {
		s.logger.Log("err", err)
		return ctx, nil, err
	}

	if len(mdHeader) > 0 {
		if err = grpc.SendHeader(ctx, mdHeader); err != nil {
			s.logger.Log("err", err)
			return ctx, nil, err
		}
	}

	if len(mdTrailer) > 0 {
		if err = grpc.SetTrailer(ctx, mdTrailer); err != nil {
			s.logger.Log("err", err)
			return ctx, nil, err
		}
	}

	return ctx, grpcResp, nil
}
