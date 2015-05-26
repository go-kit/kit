package http

import (
	"bytes"
	"net/http"
	"net/url"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/transport/codec"
)

type httpClient struct {
	*url.URL
	codec.Codec
	method string
	*http.Client
	before       []BeforeFunc
	makeResponse func() interface{}
}

// NewClient returns a client endpoint for a remote service. addr must be a
// valid, parseable URL, including the scheme and path.
func NewClient(addr string, cdc codec.Codec, makeResponse func() interface{}, options ...ClientOption) endpoint.Endpoint {
	u, err := url.Parse(addr)
	if err != nil {
		panic(err)
	}
	c := httpClient{
		URL:          u,
		Codec:        cdc,
		method:       "GET",
		Client:       http.DefaultClient,
		makeResponse: makeResponse,
	}
	for _, option := range options {
		option(&c)
	}
	return c.endpoint
}

func (c httpClient) endpoint(ctx context.Context, request interface{}) (interface{}, error) {
	var buf bytes.Buffer
	if err := c.Codec.Encode(&buf, request); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(c.method, c.URL.String(), &buf)
	if err != nil {
		return nil, err
	}

	for _, f := range c.before {
		f(ctx, req)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	response := c.makeResponse()
	ctx, err = c.Codec.Decode(ctx, resp.Body, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// ClientOption sets a parameter for the HTTP client.
type ClientOption func(*httpClient)

// ClientBefore adds pre-invocation BeforeFuncs to the HTTP client.
func ClientBefore(funcs ...BeforeFunc) ClientOption {
	return func(c *httpClient) { c.before = append(c.before, funcs...) }
}

// ClientMethod sets the method used to invoke the RPC. By default, it's GET.
func ClientMethod(method string) ClientOption {
	return func(c *httpClient) { c.method = method }
}

// SetClient sets the HTTP client struct used to invoke the RPC. By default,
// it's http.DefaultClient.
func SetClient(client *http.Client) ClientOption {
	return func(c *httpClient) { c.Client = client }
}
