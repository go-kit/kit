package http

import (
	"net/http"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

// Server wraps an endpoint and implements http.Handler.
type Server struct {
	ctx          context.Context
	e            endpoint.Endpoint
	dec          DecodeRequestFunc
	enc          EncodeResponseFunc
	before       []RequestFunc
	after        []ResponseFunc
	errorEncoder func(w http.ResponseWriter, err error)
	logger       log.Logger
}

// NewServer constructs a new server, which implements http.Server and wraps
// the provided endpoint.
func NewServer(
	ctx context.Context,
	e endpoint.Endpoint,
	dec DecodeRequestFunc,
	enc EncodeResponseFunc,
	options ...ServerOption,
) *Server {
	s := &Server{
		ctx:          ctx,
		e:            e,
		dec:          dec,
		enc:          enc,
		errorEncoder: defaultErrorEncoder,
		logger:       log.NewNopLogger(),
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

// ServerErrorEncoder is used to encode errors to the http.ResponseWriter
// whenever they're encountered in the processing of a request. Clients can
// use this to provide custom error formatting and response codes. By default,
// errors will be written as plain text with an appropriate, if generic,
// status code.
func ServerErrorEncoder(f func(w http.ResponseWriter, err error)) ServerOption {
	return func(s *Server) { s.errorEncoder = f }
}

// ServerErrorLogger is used to log non-terminal errors. By default, no errors
// are logged.
func ServerErrorLogger(logger log.Logger) ServerOption {
	return func(s *Server) { s.logger = logger }
}

// ServeHTTP implements http.Handler.
func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(s.ctx)
	defer cancel()

	for _, f := range s.before {
		ctx = f(ctx, r)
	}

	request, err := s.dec(r)
	if err != nil {
		s.logger.Log("err", err)
		s.errorEncoder(w, BadRequestError{err})
		return
	}

	response, err := s.e(ctx, request)
	if err != nil {
		s.logger.Log("err", err)
		s.errorEncoder(w, err)
		return
	}

	for _, f := range s.after {
		f(ctx, w)
	}

	if err := s.enc(w, response); err != nil {
		s.logger.Log("err", err)
		s.errorEncoder(w, err)
		return
	}
}

func defaultErrorEncoder(w http.ResponseWriter, err error) {
	switch err.(type) {
	case BadRequestError:
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// BadRequestError is an error in decoding the request.
type BadRequestError struct {
	Err error
}

// Error implements the error interface.
func (err BadRequestError) Error() string {
	return err.Err.Error()
}
