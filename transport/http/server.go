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
}

// ServeHTTP implements http.Handler.
func (b Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(b.Context)
	defer cancel()

	for _, f := range b.Before {
		ctx = f(ctx, r)
	}

	request, err := b.DecodeRequestFunc(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := r.Body.Close(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := b.Endpoint(ctx, request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, f := range b.After {
		f(ctx, w)
	}

	if err := b.EncodeResponseFunc(w, response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
