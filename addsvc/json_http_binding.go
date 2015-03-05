package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/peterbourgon/gokit/transport/codec"

	"github.com/peterbourgon/gokit/server"
	"golang.org/x/net/context"
)

type jsonCodec struct{}

func (jsonCodec) Decode(_ context.Context, r io.Reader) (server.Request, error) {
	var req request
	err := json.NewDecoder(r).Decode(req)
	return req, err
}

func (jsonCodec) Encode(w io.Writer, resp server.Response) error {
	return json.NewEncoder(w).Encode(resp)
}

type httpBinding struct {
	context.Context
	codec.Codec
	server.Endpoint
}

func (b httpBinding) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// If the context is canceled, we should not perform work.
	select {
	case <-b.Context.Done():
		http.Error(w, "context is canceled", http.StatusServiceUnavailable)
		return
	default:
	}

	// Generate a context for this request.
	// TODO read headers to determine what kind of context to create
	ctx, cancel := context.WithCancel(b.Context)
	defer cancel()

	// Perform HTTP-specific context amendments.
	// TODO extract e.g. trace ID

	// Decode request.
	req, err := b.Codec.Decode(ctx, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Execute RPC.
	resp, err := b.Endpoint(ctx, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Encode response.
	if err := b.Codec.Encode(w, resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
