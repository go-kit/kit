package twirp

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

// Client wraps a Twirp client and provides a method that implements endpoint.Endpoint.
type Client struct {
	rpcFn     endpoint.Endpoint
	enc       EncodeRequestFunc
	dec       DecodeResponseFunc
	before    []ClientRequestFunc
	after     []ClientResponseFunc
	finalizer ClientFinalizerFunc
}

// NewClient constructs a usable Client for a single remote method.
func NewClient(
	rpcFn endpoint.Endpoint,
	enc EncodeRequestFunc,
	dec DecodeResponseFunc,
	options ...ClientOption,
) *Client {
	c := &Client{
		rpcFn:  rpcFn,
		enc:    enc,
		dec:    dec,
		before: []ClientRequestFunc{},
		after:  []ClientResponseFunc{},
	}
	for _, option := range options {
		option(c)
	}
	return c
}

// ClientOption sets an optional parameter for clients.
type ClientOption func(*Client)

// ClientBefore sets the ClientRequestFunc that are applied to the outgoing
// request before it's invoked.
func ClientBefore(before ...ClientRequestFunc) ClientOption {
	return func(c *Client) { c.before = append(c.before, before...) }
}

// ClientAfter sets the ClientResponseFuncs applied to the incoming
// request prior to it being decoded. This is useful for obtaining anything off
// of the response and adding onto the context prior to decoding.
func ClientAfter(after ...ClientResponseFunc) ClientOption {
	return func(c *Client) { c.after = append(c.after, after...) }
}

// ClientFinalizer is executed at the end of every request.
// By default, no finalizer is registered.
func ClientFinalizer(f ClientFinalizerFunc) ClientOption {
	return func(s *Client) { s.finalizer = f }
}

// Endpoint returns a usable endpoint that invokes the remote endpoint.
func (c Client) Endpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		var (
			req interface{}
			err error
		)

		// Process ClientFinalizers
		if c.finalizer != nil {
			defer func() {
				c.finalizer(ctx, err)
			}()
		}

		// Encode
		req, err = c.enc(ctx, request)
		if err != nil {
			return nil, err
		}

		// Process ClientRequestFunctions
		for _, f := range c.before {
			ctx, err = f(ctx)
			if err != nil {
				return nil, err
			}
		}

		// Call the actual RPC method
		resp, err := c.rpcFn(ctx, req)
		if err != nil {
			return nil, err
		}

		// Process ClientResponseFunctions
		for _, f := range c.after {
			ctx, err = f(ctx)
			if err != nil {
				return nil, err
			}
		}

		// Decode
		response, err := c.dec(ctx, resp)
		if err != nil {
			return nil, err
		}

		return response, nil
	}
}

// ClientFinalizerFunc can be used to perform work at the end of a client
// request, after the response is returned. The principal
// intended use is for error logging. Note: err may be nil.
// There maybe also no additional response parameters depending on when
// an error occurs.
type ClientFinalizerFunc func(ctx context.Context, err error)
