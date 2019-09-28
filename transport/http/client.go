package http

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/go-kit/kit/endpoint"
)

// NewClientEndpoint returns a usable endpoint that invokes the remote endpoint.
func NewClientEndpoint(
	method string,
	tgt *url.URL,
	enc EncodeClientRequestFunc,
	dec DecodeResponseFunc,
	options ...ClientEndpointOption,
) endpoint.Endpoint {
	opts := &clientEndpointOpts{
		client:         http.DefaultClient,
		before:         []RequestFunc{},
		after:          []ClientResponseFunc{},
		bufferedStream: false,
	}
	for _, option := range options {
		option(opts)
	}
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		ctx, cancel := context.WithCancel(ctx)

		var (
			resp *http.Response
			err  error
		)
		if opts.finalizer != nil {
			defer func() {
				if resp != nil {
					ctx = context.WithValue(ctx, ContextKeyResponseHeaders, resp.Header)
					ctx = context.WithValue(ctx, ContextKeyResponseSize, resp.ContentLength)
				}
				for _, f := range opts.finalizer {
					f(ctx, err)
				}
			}()
		}

		req, err := enc(ctx, method, tgt.String(), request)
		if err != nil {
			cancel()
			return nil, err
		}
		// If user didn't bother to create a req object we have to set a default value for it ourselves.
		if req == nil {
			req, err = http.NewRequest(method, tgt.String(), nil)
			if err != nil {
				cancel()
				return nil, err
			}
		}

		for _, f := range opts.before {
			ctx = f(ctx, req)
		}

		resp, err = opts.client.Do(req.WithContext(ctx))

		if err != nil {
			cancel()
			return nil, err
		}

		// If we expect a buffered stream, we don't cancel the context when the endpoint returns.
		// Instead, we should call the cancel func when closing the response body.
		if opts.bufferedStream {
			resp.Body = bodyWithCancel{ReadCloser: resp.Body, cancel: cancel}
		} else {
			defer resp.Body.Close()
			defer cancel()
		}

		for _, f := range opts.after {
			ctx = f(ctx, resp)
		}

		response, err := dec(ctx, resp)
		if err != nil {
			return nil, err
		}

		return response, nil
	}
}

// ClientEndpointOption sets an optional parameter for client endpoint.
type ClientEndpointOption func(*clientEndpointOpts)

type clientEndpointOpts struct {
	client         HTTPClient
	before         []RequestFunc
	after          []ClientResponseFunc
	finalizer      []ClientFinalizerFunc
	bufferedStream bool
}

// ClientEndpointSetClient sets the underlying HTTP client used for requests.
// By default, http.DefaultClient is used.
func ClientEndpointSetClient(client HTTPClient) ClientEndpointOption {
	return func(e *clientEndpointOpts) { e.client = client }
}

// ClientEndpointBefore sets the RequestFuncs that are applied to the outgoing HTTP
// request before it's invoked.
func ClientEndpointBefore(before ...RequestFunc) ClientEndpointOption {
	return func(e *clientEndpointOpts) { e.before = append(e.before, before...) }
}

// ClientEndpointAfter sets the ClientResponseFuncs applied to the incoming HTTP
// request prior to it being decoded. This is useful for obtaining anything off
// of the response and adding onto the context prior to decoding.
func ClientEndpointAfter(after ...ClientResponseFunc) ClientEndpointOption {
	return func(e *clientEndpointOpts) { e.after = append(e.after, after...) }
}

// ClientEndpointFinalizer is executed at the end of every HTTP request.
// By default, no finalizer is registered.
func ClientEndpointFinalizer(f ...ClientFinalizerFunc) ClientEndpointOption {
	return func(e *clientEndpointOpts) { e.finalizer = append(e.finalizer, f...) }
}

// ClientEndpointBufferedStream sets whether the Response.Body is left open, allowing it
// to be read from later. Useful for transporting a file as a buffered stream.
// That body has to be Closed to propery end the request.
func ClientEndpointBufferedStream(buffered bool) ClientEndpointOption {
	return func(e *clientEndpointOpts) { e.bufferedStream = buffered }
}

// HTTPClient is an interface that models *http.Client.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// bodyWithCancel is a wrapper for an io.ReadCloser with also a
// cancel function which is called when the Close is used
type bodyWithCancel struct {
	io.ReadCloser

	cancel context.CancelFunc
}

func (bwc bodyWithCancel) Close() error {
	bwc.ReadCloser.Close()
	bwc.cancel()
	return nil
}

// ClientFinalizerFunc can be used to perform work at the end of a client HTTP
// request, after the response is returned. The principal
// intended use is for error logging. Additional response parameters are
// provided in the context under keys with the ContextKeyResponse prefix.
// Note: err may be nil. There maybe also no additional response parameters
// depending on when an error occurs.
type ClientFinalizerFunc func(ctx context.Context, err error)

// EncodeJSONClientRequest is an EncodeRequestFunc that serializes the request as a
// JSON object to the Request body. Many JSON-over-HTTP services can use it as
// a sensible default. If the request implements Headerer, the provided headers
// will be applied to the request.
func EncodeJSONClientRequest(c context.Context, method, url string, request interface{}) (*http.Request, error) {
	r, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	r.Header.Set("Content-Type", "application/json; charset=utf-8")
	if headerer, ok := request.(Headerer); ok {
		for k := range headerer.Headers() {
			r.Header.Set(k, headerer.Headers().Get(k))
		}
	}
	var b bytes.Buffer
	r.Body = ioutil.NopCloser(&b)
	if err := json.NewEncoder(&b).Encode(request); err != nil {
		return nil, err
	}
	return r, nil
}

// EncodeXMLClientRequest is an EncodeRequestFunc that serializes the request as a
// XML object to the Request body. If the request implements Headerer,
// the provided headers will be applied to the request.
func EncodeXMLClientRequest(c context.Context, method, url string, request interface{}) (*http.Request, error) {
	r, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	r.Header.Set("Content-Type", "text/xml; charset=utf-8")
	if headerer, ok := request.(Headerer); ok {
		for k := range headerer.Headers() {
			r.Header.Set(k, headerer.Headers().Get(k))
		}
	}
	var b bytes.Buffer
	r.Body = ioutil.NopCloser(&b)
	if err := xml.NewEncoder(&b).Encode(request); err != nil {
		return nil, err
	}
	return r, nil
}
