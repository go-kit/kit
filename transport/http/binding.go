package http

import (
	"net/http"

	"golang.org/x/net/context"

	"github.com/peterbourgon/gokit/server"
	"github.com/peterbourgon/gokit/transport/codec"
)

// BindingOption sets a parameter for the binding.
type BindingOption func(*binding)

// Before adds pre-RPC BeforeFuncs to the binding.
func Before(funcs ...BeforeFunc) BindingOption {
	return func(b *binding) { b.before = append(b.before, funcs...) }
}

// After adds post-RPC AfterFuncs to the binding.
func After(funcs ...AfterFunc) BindingOption {
	return func(b *binding) { b.after = append(b.after, funcs...) }
}

type binding struct {
	context.Context
	codec.Codec
	server.Endpoint
	before []BeforeFunc
	after  []AfterFunc
}

// NewBinding returns an HTTP handler that wraps the given endpoint.
func NewBinding(ctx context.Context, cdc codec.Codec, endpoint server.Endpoint, options ...BindingOption) http.Handler {
	b := &binding{
		Context:  ctx,
		Codec:    cdc,
		Endpoint: endpoint,
	}
	for _, option := range options {
		option(b)
	}
	return b
}

func (b *binding) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Per-request context.
	ctx, cancel := context.WithCancel(b.Context)
	defer cancel()

	// Prepare the RPC's context with details from the request.
	for _, f := range b.before {
		ctx = f(ctx, r)
	}

	// Decode request.
	req, ctx, err := b.Codec.Decode(ctx, r.Body)
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

	// Prepare the ResponseWriter.
	for _, f := range b.after {
		f(ctx, w)
	}

	// Encode response.
	if err := b.Codec.Encode(w, resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
