package jsonrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
)

// Server wraps an endpoint and implements http.Handler.
type Server struct {
	ctx       context.Context
	ecm       EndpointCodecMap
	before    []httptransport.RequestFunc
	after     []httptransport.ServerResponseFunc
	finalizer httptransport.ServerFinalizerFunc
	logger    log.Logger
}

// NewServer constructs a new server, which implements http.Server.
func NewServer(
	ctx context.Context,
	ecm EndpointCodecMap,
	options ...ServerOption,
) *Server {
	s := &Server{
		ctx:    ctx,
		ecm:    ecm,
		logger: log.NewNopLogger(),
	}
	for _, option := range options {
		option(s)
	}
	return s
}

// EndpointCodec defines and Endpoint and its associated codecs
type EndpointCodec struct {
	Endpoint endpoint.Endpoint
	Decode   DecodeRequestFunc
	Encode   EncodeResponseFunc
}

// EndpointCodecMap maps the Request.Method to the proper EndpointCodec
type EndpointCodecMap map[string]EndpointCodec

// ServerOption sets an optional parameter for servers.
type ServerOption func(*Server)

// ServerBefore functions are executed on the HTTP request object before the
// request is decoded.
func ServerBefore(before ...httptransport.RequestFunc) ServerOption {
	return func(s *Server) { s.before = append(s.before, before...) }
}

// ServerAfter functions are executed on the HTTP response writer after the
// endpoint is invoked, but before anything is written to the client.
func ServerAfter(after ...httptransport.ServerResponseFunc) ServerOption {
	return func(s *Server) { s.after = append(s.after, after...) }
}

// ServerErrorLogger is used to log non-terminal errors. By default, no errors
// are logged. This is intended as a diagnostic measure. Finer-grained control
// of error handling, including logging in more detail, should be performed in a
// custom ServerErrorEncoder or ServerFinalizer, both of which have access to
// the context.
func ServerErrorLogger(logger log.Logger) ServerOption {
	return func(s *Server) { s.logger = logger }
}

// ServerFinalizer is executed at the end of every HTTP request.
// By default, no finalizer is registered.
func ServerFinalizer(f httptransport.ServerFinalizerFunc) ServerOption {
	return func(s *Server) { s.finalizer = f }
}

// ServeHTTP implements http.Handler.
func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "405 must POST\n")
		return
	}
	ctx := s.ctx

	if s.finalizer != nil {
		iw := &interceptingWriter{w, http.StatusOK}
		defer func() { s.finalizer(ctx, iw.code, r) }()
		w = iw
	}

	for _, f := range s.before {
		ctx = f(ctx, r)
	}

	// Decode the body into an  object
	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		s.logger.Log("err", err)
		rpcErrorEncoder(ctx, err, w)
		return
	}

	// Get the endpoint and codecs from the map using the method
	// defined in the JSON  object
	ecm, ok := s.ecm[req.Method]
	if !ok {
		err := methodNotFoundError(fmt.Sprintf("Method %s was not found.", req.Method))
		s.logger.Log("err", err)
		rpcErrorEncoder(ctx, err, w)
		return
	}

	// Decode the JSON  "params"
	reqParams, err := ecm.Decode(ctx, req.Params)
	if err != nil {
		s.logger.Log("err", err)
		rpcErrorEncoder(ctx, err, w)
		return
	}

	// Call the Endpoint with the params
	response, err := ecm.Endpoint(ctx, reqParams)
	if err != nil {
		s.logger.Log("err", err)
		rpcErrorEncoder(ctx, err, w)
		return
	}

	for _, f := range s.after {
		ctx = f(ctx, w)
	}

	res := Response{}

	// Encode the response from the Endpoint
	resParams, err := ecm.Encode(ctx, response)
	if err != nil {
		s.logger.Log("err", err)
		rpcErrorEncoder(ctx, err, w)
		return
	}

	res.Result = resParams

	json.NewEncoder(w).Encode(res)
}

// ErrorEncoder writes the error to the ResponseWriter, by default a
// content type of text/plain, a body of the plain text of the error, and a
// status code of 500. If the error implements Headerer, the provided headers
// will be applied to the response. If the error implements json.Marshaler, and
// the marshaling succeeds, a content type of application/json and the JSON
// encoded form of the error will be used. If the error implements StatusCoder,
// the provided StatusCode will be used instead of 500.
func rpcErrorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", ContentType)
	if headerer, ok := err.(httptransport.Headerer); ok {
		for k := range headerer.Headers() {
			w.Header().Set(k, headerer.Headers().Get(k))
		}
	}

	e := Error{
		Code:    InternalError,
		Message: err.Error(),
	}
	if sc, ok := err.(ErrorCoder); ok {
		e.Code = sc.ErrorCode()
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{
		JSONRPC: Version,
		Error:   &e,
	})
}

// ErrorCoder is checked by DefaultErrorEncoder. If an error value implements
// ErrorCoder, the Error will be used when encoding the error. By default,
// InternalError (-32603) is used.
type ErrorCoder interface {
	ErrorCode() int
}

type interceptingWriter struct {
	http.ResponseWriter
	code int
}

// WriteHeader may not be explicitly called, so care must be taken to
// initialize w.code to its default value of http.StatusOK.
func (w *interceptingWriter) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}
