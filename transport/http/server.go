package http

import (
	"net/http"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
)

// Server wraps an endpoint and implements http.Handler.
type Server struct {
	// A background context must be provided.
	context.Context

	// The endpoint that will be invoked.
	endpoint.Endpoint

	// DecodeRequestFunc must be provided.
	DecodeRequestFunc

	// EncodeResponseFunc must be provided.
	EncodeResponseFunc

	// Before functions are executed on the HTTP request object before the
	// request is decoded.
	Before []RequestFunc

	// After functions are executed on the HTTP response writer after the
	// endpoint is invoked, but before anything is written to the client.
	After []ResponseFunc

	// ErrorEncoder is used to encode errors to the http.ResponseWriter
	// whenever they're encountered in the processing of a request. Clients
	// can use this to provide custom error formatting and response codes. If
	// ErrorEncoder is nil, the error will be written as plain text with
	// an appropriate, if generic, status code.
	ErrorEncoder func(w http.ResponseWriter, err error)
}

// ServeHTTP implements http.Handler.
func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.ErrorEncoder == nil {
		s.ErrorEncoder = defaultErrorEncoder
	}

	ctx, cancel := context.WithCancel(s.Context)
	defer cancel()

	for _, f := range s.Before {
		ctx = f(ctx, r)
	}

	request, err := s.DecodeRequestFunc(r)
	if err != nil {
		s.ErrorEncoder(w, badRequestError{err})
		return
	}

	response, err := s.Endpoint(ctx, request)
	if err != nil {
		s.ErrorEncoder(w, err)
		return
	}

	for _, f := range s.After {
		f(ctx, w)
	}

	if err := s.EncodeResponseFunc(w, response); err != nil {
		s.ErrorEncoder(w, err)
		return
	}
}

func defaultErrorEncoder(w http.ResponseWriter, err error) {
	switch err.(type) {
	case badRequestError:
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type badRequestError struct{ error }
