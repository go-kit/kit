package http

import (
	"net/http"
	"net/url"

	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"

	"github.com/go-kit/kit/endpoint"
)

// Client wraps a URL and provides a method that implements endpoint.Endpoint.
type Client struct {
	client         *http.Client
	method         string
	tgt            *url.URL
	enc            EncodeRequestFunc
	dec            DecodeResponseFunc
	before         []RequestFunc
	bufferedStream bool
}

// NewClient constructs a usable Client for a single remote endpoint.
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
		bufferedStream: false,
	}
	for _, option := range options {
		option(c)
	}
	return c
}

// ClientOption sets an optional parameter for clients.
type ClientOption func(*Client)

// SetClient sets the underlying HTTP client used for requests.
// By default, http.DefaultClient is used.
func SetClient(client *http.Client) ClientOption {
	return func(c *Client) { c.client = client }
}

// SetClientBefore sets the RequestFuncs that are applied to the outgoing HTTP
// request before it's invoked.
func SetClientBefore(before ...RequestFunc) ClientOption {
	return func(c *Client) { c.before = before }
}

// SetBufferedStream sets whether the Response.Body is left open, allowing it
// to be read from later. Useful for transporting a file as a buffered stream.
func SetBufferedStream(buffered bool) ClientOption {
	return func(c *Client) { c.bufferedStream = buffered }
}

// Endpoint returns a usable endpoint that will invoke the RPC specified by
// the client.
func (c Client) Endpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		req, err := http.NewRequest(c.method, c.tgt.String(), nil)
		if err != nil {
			return nil, TransportError{DomainNewRequest, err}
		}

		if err = c.enc(req, request); err != nil {
			return nil, TransportError{DomainEncode, err}
		}

		for _, f := range c.before {
			ctx = f(ctx, req)
		}

		resp, err := ctxhttp.Do(ctx, c.client, req)
		if err != nil {
			return nil, TransportError{DomainDo, err}
		}
		if !c.bufferedStream {
			defer resp.Body.Close()
		}

		response, err := c.dec(resp)
		if err != nil {
			return nil, TransportError{DomainDecode, err}
		}

		return response, nil
	}
}
