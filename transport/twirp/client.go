package twirp

import (
	"context"
	"fmt"
	"github.com/go-kit/kit/endpoint"
	"github.com/twitchtv/twirp"
	"net/http"
	"reflect"
)

// Client wraps a Twirp client and provides a method that implements endpoint.Endpoint.
type Client struct {
	client    interface{}
	method    string
	enc       EncodeRequestFunc
	dec       DecodeResponseFunc
	before    []ClientRequestFunc
	after     []ClientResponseFunc
	finalizer ClientFinalizerFunc
}

// NewClient constructs a usable Client for a single remote method.
func NewClient(
	client interface{},
	method string,
	enc EncodeRequestFunc,
	dec DecodeResponseFunc,
	options ...ClientOption,
) *Client {
	c := &Client{
		client: client,
		method: method,
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

		// Create an empty http.Header to hold the headers that we will accumulate in before functions.
		var reqHeader http.Header
		// Process ClientRequestFunctions
		for _, f := range c.before {
			ctx = f(ctx, &reqHeader)
		}

		// Tell twirp to use these headers in the request.
		ctx, err = twirp.WithHTTPRequestHeaders(ctx, reqHeader)
		if err != nil {
			return nil, err
		}

		client := reflect.ValueOf(&c.client)
		method := client.MethodByName(c.method)
		if !method.IsValid() {
			interfaceName := reflect.TypeOf(&c.client).Elem().Name()
			return nil, fmt.Errorf("Invalid method specified: %s does not have method %s", interfaceName, c.method)
		}

		args := make([]reflect.Value, 2)
		args[0] = reflect.ValueOf(ctx)
		args[1] = reflect.ValueOf(req)

		retVals := make([]reflect.Value, 2)
		retVals = method.Call(args)
		resp := retVals[0].Interface()
		err = retVals[1].Interface().(error)
		if err != nil {
			return nil, err
		}

		// Process ClientResponseFunctions
		for _, f := range c.after {
			ctx = f(ctx)
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
