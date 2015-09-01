package http

import (
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
)

// Client wraps a URL and provides a method that implements endpoint.Endpoint.
type Client struct {
	// If client is nil, http.DefaultClient will be used.
	*http.Client

	// Method must be provided.
	Method string

	// URL must be provided.
	URL *url.URL

	// EncodeRequestFunc must be provided. The HTTP request passed to the
	// EncodeRequestFunc will have a nil body.
	EncodeRequestFunc

	// DecodeResponseFunc must be provided.
	DecodeResponseFunc

	// Before functions are executed on the outgoing request after it is
	// created, but before it's sent to the HTTP client. Clients have no After
	// ResponseFuncs, as they don't work with ResponseWriters.
	Before []RequestFunc
}

// Endpoint returns a usable endpoint that will invoke the RPC specified by
// the client.
func (c Client) Endpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		req, err := http.NewRequest(c.Method, c.URL.String(), nil)
		if err != nil {
			return nil, fmt.Errorf("NewRequest: %v", err)
		}

		if err = c.EncodeRequestFunc(req, request); err != nil {
			return nil, fmt.Errorf("Encode: %v", err)
		}

		for _, f := range c.Before {
			ctx = f(ctx, req)
		}

		var resp *http.Response
		if c.Client != nil {
			resp, err = c.Client.Do(req)
		} else {
			resp, err = http.DefaultClient.Do(req)
		}
		if err != nil {
			return nil, fmt.Errorf("Do: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		response, err := c.DecodeResponseFunc(resp)
		if err != nil {
			return nil, fmt.Errorf("Decode: %v", err)
		}

		return response, nil
	}
}
