package httpjson

import (
	"net/http"
	"net/url"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/examples/addsvc/server"
	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
)

// New returns an AddService that's backed by the URL. baseurl will have its
// scheme and hostport used, but its path will be overwritten. If client is
// nil, http.DefaultClient will be used.
func New(ctx context.Context, baseurl *url.URL, logger log.Logger, c *http.Client) server.AddService {
	sumURL, err := url.Parse(baseurl.String())
	if err != nil {
		panic(err)
	}
	sumURL.Path = "/sum"

	concatURL, err := url.Parse(baseurl.String())
	if err != nil {
		panic(err)
	}
	concatURL.Path = "/concat"

	return client{
		Context: ctx,
		Logger:  logger,
		sum: httptransport.NewClient(
			"GET",
			sumURL,
			server.EncodeSumRequest,
			server.DecodeSumResponse,
			httptransport.SetClient(c),
		).Endpoint(),
		concat: httptransport.NewClient(
			"GET",
			concatURL,
			server.EncodeConcatRequest,
			server.DecodeConcatResponse,
			httptransport.SetClient(c),
		).Endpoint(),
	}
}

type client struct {
	context.Context
	log.Logger
	sum    endpoint.Endpoint
	concat endpoint.Endpoint
}

func (c client) Sum(a, b int) int {
	response, err := c.sum(c.Context, server.SumRequest{A: a, B: b})
	if err != nil {
		c.Logger.Log("err", err)
		return 0
	}
	return response.(server.SumResponse).V
}

func (c client) Concat(a, b string) string {
	response, err := c.concat(c.Context, server.ConcatRequest{A: a, B: b})
	if err != nil {
		c.Logger.Log("err", err)
		return ""
	}
	return response.(server.ConcatResponse).V
}
