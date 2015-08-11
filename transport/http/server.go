package http

import (
	"net/http"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
)

// Server wraps an endpoint and implements http.Handler.
type Server struct {
	context.Context
	endpoint.Endpoint
	DecodeFunc
	EncodeFunc
	Before []RequestFunc
	After  []ResponseFunc
}

// ServeHTTP implements http.Handler.
func (b Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(b.Context)
	defer cancel()

	for _, f := range b.Before {
		ctx = f(ctx, r)
	}

	request, err := b.DecodeFunc(r.Body)
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

	if err := b.EncodeFunc(w, response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
