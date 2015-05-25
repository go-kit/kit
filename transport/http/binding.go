package http

import (
	"net/http"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/transport/codec"
)

type binding struct {
	context.Context
	makeRequest func() interface{}
	codec.Codec
	endpoint.Endpoint
	before []BeforeFunc
	after  []AfterFunc
}

// NewBinding returns an HTTP handler that wraps the given endpoint.
func NewBinding(ctx context.Context, makeRequest func() interface{}, cdc codec.Codec, e endpoint.Endpoint, options ...BindingOption) http.Handler {
	b := &binding{
		Context:     ctx,
		makeRequest: makeRequest,
		Codec:       cdc,
		Endpoint:    e,
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
	req := b.makeRequest()
	ctx, err := b.Codec.Decode(ctx, r.Body, req)
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

// BindingOption sets a parameter for the HTTP binding.
type BindingOption func(*binding)

// BindingBefore adds pre-RPC BeforeFuncs to the HTTP binding.
func BindingBefore(funcs ...BeforeFunc) BindingOption {
	return func(b *binding) { b.before = append(b.before, funcs...) }
}

// BindingAfter adds post-RPC AfterFuncs to the HTTP binding.
func BindingAfter(funcs ...AfterFunc) BindingOption {
	return func(b *binding) { b.after = append(b.after, funcs...) }
}
