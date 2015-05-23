package main

import (
	"bytes"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-kit/kit/log"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/client"
	"github.com/go-kit/kit/transport/codec"
	httptransport "github.com/go-kit/kit/transport/http"
)

// Add is the abstract definition of what this service does. It could easily
// be an interface type with multiple methods, in which case each method would
// be an endpoint.
type Add func(context.Context, int64, int64) int64

func pureAdd(_ context.Context, a, b int64) int64 { return a + b }

func proxyAdd(e client.Endpoint) Add {
	return func(ctx context.Context, a, b int64) int64 {
		resp, err := e(ctx, &addRequest{a, b})
		if err != nil {
			log.DefaultLogger.Log("err", err)
			return 0
		}
		addResp, ok := resp.(*addResponse)
		if !ok {
			log.DefaultLogger.Log("err", client.ErrBadCast)
			return 0
		}
		return addResp.V
	}
}

type httpClient struct {
	*url.URL
	codec.Codec
	method string
	*http.Client
	before       []httptransport.BeforeFunc
	makeResponse func() interface{}
}

// TODO this needs to go to package client
func newHTTPClient(addr string, cdc codec.Codec, makeResponse func() interface{}, options ...httpClientOption) client.Endpoint {
	if !strings.HasPrefix(addr, "http") {
		addr = "http://" + addr
	}
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

type httpClientOption func(*httpClient)

func before(f httptransport.BeforeFunc) httpClientOption {
	return func(c *httpClient) { c.before = append(c.before, f) }
}
