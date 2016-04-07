package httpjson

import (
	"io"
	"net/url"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/examples/addsvc/server"
	httptransport "github.com/go-kit/kit/transport/http"
)

// SumEndpointFactory transforms a http url into an Endpoint.
// The path of the url is reset to /sum.
func SumEndpointFactory(instance string) (endpoint.Endpoint, io.Closer, error) {
	sumURL, err := url.Parse(instance)
	if err != nil {
		return nil, nil, err
	}
	sumURL.Path = "/sum"

	client := httptransport.NewClient(
		"GET",
		sumURL,
		server.EncodeSumRequest,
		server.DecodeSumResponse,
		httptransport.SetClient(nil),
	)

	return client.Endpoint(), nil, nil
}

// ConcatEndpointFactory transforms a http url into an Endpoint.
// The path of the url is reset to /concat.
func ConcatEndpointFactory(instance string) (endpoint.Endpoint, io.Closer, error) {
	concatURL, err := url.Parse(instance)
	if err != nil {
		return nil, nil, err
	}
	concatURL.Path = "/concat"

	client := httptransport.NewClient(
		"GET",
		concatURL,
		server.EncodeConcatRequest,
		server.DecodeConcatResponse,
		httptransport.SetClient(nil),
	)

	return client.Endpoint(), nil, nil
}
