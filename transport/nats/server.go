package nats

import (
	"context"

	"fmt"
	"os"

	"github.com/go-kit/kit/endpoint"
	"github.com/nats-io/go-nats"
	log "github.com/sirupsen/logrus"

)

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
		logger: *log.New(),
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

// MsgHandler implements the MsgHandler type.
func (s Server) MsgHandler(msg *nats.Msg) {

	urls := fmt.Sprint(os.Getenv("NATS_SERVER"), ", ", fmt.Sprintf("nats://%v:%v", os.Getenv("NATS_SERVICE_HOST"), os.Getenv("NATS_SERVICE_PORT")))

	s.logger.Info(urls)

	nc, err := nats.Connect(urls)
	if err != nil {
		s.logger.Error("Can't connect: %v\n", err)
	}

	defer nc.Close()

	// Non-nil non empty context to take the place of the first context in th chain of handling.
	ctx := context.TODO()

	request, err := s.dec(ctx, msg)
	if err != nil {
		s.logger.Error("err", err)
		return
	}

	response, err := s.e(ctx, request)
	if err != nil {
		s.logger.Error("err", err)
		return
	}

	payload, err := s.enc(ctx, response)
	if err != nil {
		s.logger.Error("err", err)
		return
	}

	nc.Publish(msg.Reply, payload)
}
