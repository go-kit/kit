package httpjson

import (
	"io"
	"net/url"

	"github.com/opentracing/opentracing-go"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/examples/addsvc/server"
	"github.com/go-kit/kit/loadbalancer"
	"github.com/go-kit/kit/log"
	kitot "github.com/go-kit/kit/tracing/opentracing"
	httptransport "github.com/go-kit/kit/transport/http"
)

// MakeSumEndpointFactory generates a Factory that transforms an http url into
// an Endpoint.
//
// The path of the url is reset to /sum.
func MakeSumEndpointFactory(tracer opentracing.Tracer, tracingLogger log.Logger) loadbalancer.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
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
			httptransport.SetClientBefore(kitot.ToHTTPRequest(tracer, tracingLogger)),
		)

		return client.Endpoint(), nil, nil
	}
}

// MakeConcatEndpointFactory generates a Factory that transforms an http url
// into an Endpoint.
//
// The path of the url is reset to /concat.
func MakeConcatEndpointFactory(tracer opentracing.Tracer, tracingLogger log.Logger) loadbalancer.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
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
			httptransport.SetClientBefore(kitot.ToHTTPRequest(tracer, tracingLogger)),
		)

		return client.Endpoint(), nil, nil
	}
}
