package ws

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"github.com/go-kit/kit/log"
)

const (
	DefaultBufferSize = 1024
)

// Server wraps an endpoint and implements http.Handler.
type Server struct {
	upgrader websocket.Upgrader
	scm      SubprotocolCodecMap
	ecm      EndpointCodecMap
	// before       []RequestFunc
	// after        []ServerResponseFunc
	errorEncoder ErrorEncoder
	finalizer    ServerFinalizerFunc
	logger       log.Logger
}

// NewServer constructs a new server, which implements http.Server.
func NewServer(
	upgrader websocket.Upgrader,
	scm SubprotocolCodecMap,
	ecm EndpointCodecMap,
	options ...ServerOption,
) *Server {
	s := &Server{
		upgrader:     upgrader,
		scm:          scm,
		ecm:          ecm,
		errorEncoder: DefaultErrorEncoder,
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
// func ServerBefore(before ...RequestFunc) ServerOption {
// 	return func(s *Server) { s.before = append(s.before, before...) }
// }

// ServerAfter functions are executed on the HTTP response writer after the
// endpoint is invoked, but before anything is written to the client.
// func ServerAfter(after ...ServerResponseFunc) ServerOption {
// 	return func(s *Server) { s.after = append(s.after, after...) }
// }

// ServerErrorEncoder is used to encode errors to the http.ResponseWriter
// whenever they're encountered in the processing of a request. Clients can
// use this to provide custom error formatting and response codes. By default,
// errors will be written with the DefaultErrorEncoder.
func ServerErrorEncoder(ee ErrorEncoder) ServerOption {
	return func(s *Server) { s.errorEncoder = ee }
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
func ServerFinalizer(f ServerFinalizerFunc) ServerOption {
	return func(s *Server) { s.finalizer = f }
}

// ServeHTTP implements http.Handler.
func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Upgrade the incoming HTTP request to a WebSocket connection
	wsconn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Log("err", err)
		return
	}

	// Extract the negotiated WebSocket Subprotocol and look it up
	// in the SubprotocolCodecMap
	subprotocol := wsconn.Subprotocol()
	sc, ok := s.scm[subprotocol]
	if !ok {
		s.logger.Log("err", fmt.Sprintf("subprotocol '%s' not supported", subprotocol))
		wsconn.Close()
		return
	}

	// Handle the WebSocket in a separate go routine with the WebSocket connection itself
	// and the proper SubprotocolCodec
	go s.serveWS(ctx, sc, wsconn)
}

func (s Server) serveWS(ctx context.Context, sc SubprotocolCodec, wsconn *websocket.Conn) {
	// Loop forever until error or close
	for {
		// Get the next Reader from the WebSocket connection
		msgtype, reader, err := wsconn.NextReader()
		if err != nil {
			s.logger.Log("err", err)
			s.errorEncoder(ctx, err, wsconn)
			break
		}

		// Use the SubprotocolCodec to decode the contents
		// of the Reader to extract the Endpoint method
		// to invoke.
		method, reader, err := sc.Decode(ctx, reader)
		if err != nil {
			s.logger.Log("err", err)
			s.errorEncoder(ctx, err, wsconn)
			break
		}

		// Find the method in the EndpointCodecMap
		e, ok := s.ecm[method]
		if !ok {
			err := methodNotFoundError(fmt.Sprintf("Method %s was not found.", method))
			s.logger.Log("err", err)
			s.errorEncoder(ctx, err, wsconn)
			return
		}

		// Use the EndpointCodec to decode the rests
		// of the contents of the reader.
		req, err := e.Decode(ctx, reader)
		fmt.Println(req, err)
		if err != nil {
			s.logger.Log("err", err)
			s.errorEncoder(ctx, err, wsconn)
			break
		}

		// Invoke the Endpoint with the concrete request object
		es, err := e.Endpoint(ctx, req)
		fmt.Println(es, err)
		if err != nil {
			s.logger.Log("err", err)
			s.errorEncoder(ctx, err, wsconn)
			break
		}

		// Set up a Buffer and encode the results of the Endpoint
		// into the Buffer.
		var buf bytes.Buffer
		err = e.Encode(ctx, &buf, es)
		if err != nil {
			s.logger.Log("err", err)
			s.errorEncoder(ctx, err, wsconn)
			break
		}

		// Get a WebSocket Writer to respond to the client
		// with the result of the Endpoint
		writer, err := wsconn.NextWriter(msgtype)
		if err != nil {
			s.logger.Log("err", err)
			s.errorEncoder(ctx, err, wsconn)
			break
		}

		// Encode the results of the Endpoint with the
		// SubprotocolCodec and write the results to the
		// WebSocket writer, close the writer to flush
		// the results to the client.
		err = sc.Encode(ctx, method, writer, &buf)
		if err != nil {
			s.logger.Log("err", err)
			s.errorEncoder(ctx, err, wsconn)
			break
		}
		writer.Close()
	}

	wsconn.Close()
}

// ErrorEncoder is responsible for encoding an error to the ResponseWriter.
// Users are encouraged to use custom ErrorEncoders to encode HTTP errors to
// their clients, and will likely want to pass and check for their own error
// types. See the example shipping/handling service.
type ErrorEncoder func(ctx context.Context, err error, wsconn *websocket.Conn)

// ServerFinalizerFunc can be used to perform work at the end of an HTTP
// request, after the response has been written to the client. The principal
// intended use is for request logging. In addition to the response code
// provided in the function signature, additional response parameters are
// provided in the context under keys with the ContextKeyResponse prefix.
type ServerFinalizerFunc func(ctx context.Context, code int, wsconn *websocket.Conn)

// DefaultErrorEncoder writes the error to the ResponseWriter,
// as a json-rpc error response, with an InternalError status code.
// The Error() string of the error will be used as the response error message.
// If the error implements ErrorCoder, the provided code will be set on the
// response error.
// If the error implements Headerer, the given headers will be set.
func DefaultErrorEncoder(_ context.Context, err error, wsconn *websocket.Conn) {
	e := Error{
		Code:    websocket.CloseInternalServerErr,
		Message: err.Error(),
	}
	if sc, ok := err.(ErrorCoder); ok {
		e.Code = sc.ErrorCode()
	}

	msg := websocket.FormatCloseMessage(e.Code, e.Message)
	wsconn.WriteControl(websocket.CloseMessage, msg, time.Now())
}

// ErrorCoder is checked by DefaultErrorEncoder. If an error value implements
// ErrorCoder, the integer result of ErrorCode() will be used as the JSONRPC
// error code when encoding the error.
//
// By default, InternalError (-32603) is used.
type ErrorCoder interface {
	ErrorCode() int
}

// interceptingWriter intercepts calls to WriteHeader, so that a finalizer
// can be given the correct status code.
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
