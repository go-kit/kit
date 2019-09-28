package http

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/go-kit/kit/endpoint"
)

// Client wraps a URL and provides a method that implements endpoint.Endpoint.
// Deprecated: Use NewClientEndpoint instead (for details see https://github.com/go-kit/kit/issues/796).
type Client struct {
	client         HTTPClient
	method         string
	tgt            *url.URL
	enc            EncodeRequestFunc
	dec            DecodeResponseFunc
	before         []RequestFunc
	after          []ClientResponseFunc
	finalizer      []ClientFinalizerFunc
	bufferedStream bool
}

// NewClient constructs a usable Client for a single remote method.
// Deprecated: Use NewClientEndpoint instead (for details see https://github.com/go-kit/kit/issues/796).
func NewClient(
	method string,
	tgt *url.URL,
	enc EncodeRequestFunc,
	dec DecodeResponseFunc,
	options ...ClientOption,
) *Client {
	c := &Client{
		client:         http.DefaultClient,
		method:         method,
		tgt:            tgt,
		enc:            enc,
		dec:            dec,
		before:         []RequestFunc{},
		after:          []ClientResponseFunc{},
		bufferedStream: false,
	}
	for _, option := range options {
		option(c)
	}
	return c
}

// Endpoint returns a usable endpoint that invokes the remote endpoint.
// Deprecated: Use NewClientEndpoint instead (for details see https://github.com/go-kit/kit/issues/796).
func (c Client) Endpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		ctx, cancel := context.WithCancel(ctx)

		var (
			resp *http.Response
			err  error
		)
		if c.finalizer != nil {
			defer func() {
				if resp != nil {
					ctx = context.WithValue(ctx, ContextKeyResponseHeaders, resp.Header)
					ctx = context.WithValue(ctx, ContextKeyResponseSize, resp.ContentLength)
				}
				for _, f := range c.finalizer {
					f(ctx, err)
				}
			}()
		}

		req, err := http.NewRequest(c.method, c.tgt.String(), nil)
		if err != nil {
			cancel()
			return nil, err
		}

		if err = c.enc(ctx, req, request); err != nil {
			cancel()
			return nil, err
		}

		for _, f := range c.before {
			ctx = f(ctx, req)
		}

		resp, err = c.client.Do(req.WithContext(ctx))

		if err != nil {
			cancel()
			return nil, err
		}

		// If we expect a buffered stream, we don't cancel the context when the endpoint returns.
		// Instead, we should call the cancel func when closing the response body.
		if c.bufferedStream {
			resp.Body = bodyWithCancel{ReadCloser: resp.Body, cancel: cancel}
		} else {
			defer resp.Body.Close()
			defer cancel()
		}

		for _, f := range c.after {
			ctx = f(ctx, resp)
		}

		response, err := c.dec(ctx, resp)
		if err != nil {
			return nil, err
		}

		return response, nil
	}
}

// ClientOption sets an optional parameter for clients.
// Deprecated: Use ClientEndpointOption instead (for details see https://github.com/go-kit/kit/issues/796).
type ClientOption func(*Client)

// SetClient sets the underlying HTTP client used for requests.
// By default, http.DefaultClient is used.
// Deprecated: Use ClientEndpointSetClient instead (for details see https://github.com/go-kit/kit/issues/796).
func SetClient(client HTTPClient) ClientOption {
	return func(c *Client) { c.client = client }
}

// ClientBefore sets the RequestFuncs that are applied to the outgoing HTTP
// request before it's invoked.
// Deprecated: Use ClientEndpointBefore instead (for details see https://github.com/go-kit/kit/issues/796).
func ClientBefore(before ...RequestFunc) ClientOption {
	return func(c *Client) { c.before = append(c.before, before...) }
}

// ClientAfter sets the ClientResponseFuncs applied to the incoming HTTP
// request prior to it being decoded. This is useful for obtaining anything off
// of the response and adding onto the context prior to decoding.
// Deprecated: Use ClientEndpointAfter instead (for details see https://github.com/go-kit/kit/issues/796).
func ClientAfter(after ...ClientResponseFunc) ClientOption {
	return func(c *Client) { c.after = append(c.after, after...) }
}

// ClientFinalizer is executed at the end of every HTTP request.
// By default, no finalizer is registered.
// Deprecated: Use ClientEndpointFinalizer instead (for details see https://github.com/go-kit/kit/issues/796).
func ClientFinalizer(f ...ClientFinalizerFunc) ClientOption {
	return func(s *Client) { s.finalizer = append(s.finalizer, f...) }
}

// BufferedStream sets whether the Response.Body is left open, allowing it
// to be read from later. Useful for transporting a file as a buffered stream.
// That body has to be Closed to propery end the request.
// Deprecated: Use ClientEndpointBufferedStream instead (for details see https://github.com/go-kit/kit/issues/796).
func BufferedStream(buffered bool) ClientOption {
	return func(c *Client) { c.bufferedStream = buffered }
}

// EncodeJSONRequest is an EncodeRequestFunc that serializes the request as a
// JSON object to the Request body. Many JSON-over-HTTP services can use it as
// a sensible default. If the request implements Headerer, the provided headers
// will be applied to the request.
// Deprecated: Use EncodeJSONClientRequest instead (for details see https://github.com/go-kit/kit/issues/796).
func EncodeJSONRequest(c context.Context, r *http.Request, request interface{}) error {
	r.Header.Set("Content-Type", "application/json; charset=utf-8")
	if headerer, ok := request.(Headerer); ok {
		for k := range headerer.Headers() {
			r.Header.Set(k, headerer.Headers().Get(k))
		}
	}
	var b bytes.Buffer
	r.Body = ioutil.NopCloser(&b)
	return json.NewEncoder(&b).Encode(request)
}

// EncodeXMLRequest is an EncodeRequestFunc that serializes the request as a
// XML object to the Request body. If the request implements Headerer,
// the provided headers will be applied to the request.
// Deprecated: Use EncodeXMLClientRequest instead (for details see https://github.com/go-kit/kit/issues/796).
func EncodeXMLRequest(c context.Context, r *http.Request, request interface{}) error {
	r.Header.Set("Content-Type", "text/xml; charset=utf-8")
	if headerer, ok := request.(Headerer); ok {
		for k := range headerer.Headers() {
			r.Header.Set(k, headerer.Headers().Get(k))
		}
	}
	var b bytes.Buffer
	r.Body = ioutil.NopCloser(&b)
	return xml.NewEncoder(&b).Encode(request)
}
