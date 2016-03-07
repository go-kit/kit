package booking

import (
	"net/url"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/routing"
	"github.com/marcusolsson/goddd/voyage"
	"golang.org/x/net/context"

	kithttp "github.com/go-kit/kit/transport/http"
)

// ServiceMiddleware ...
type ServiceMiddleware func(routing.Service) routing.Service

type proxyRoutingService struct {
	context.Context
	FetchRoutesEndpoint endpoint.Endpoint
	routing.Service
}

func (s proxyRoutingService) FetchRoutesForSpecification(rs cargo.RouteSpecification) []cargo.Itinerary {
	response, err := s.FetchRoutesEndpoint(s.Context, fetchRoutesRequest{
		From: string(rs.Origin),
		To:   string(rs.Destination),
	})
	if err != nil {
		return []cargo.Itinerary{}
	}

	resp := response.(fetchRoutesResponse)

	var itineraries []cargo.Itinerary
	for _, r := range resp.Paths {
		var legs []cargo.Leg
		for _, e := range r.Edges {
			legs = append(legs, cargo.Leg{
				VoyageNumber:   voyage.Number(e.Voyage),
				LoadLocation:   location.UNLocode(e.Origin),
				UnloadLocation: location.UNLocode(e.Destination),
				LoadTime:       e.Departure,
				UnloadTime:     e.Arrival,
			})
		}

		itineraries = append(itineraries, cargo.Itinerary{Legs: legs})
	}

	return itineraries
}

// ProxyingMiddleware ...
func ProxyingMiddleware(proxyURL string, ctx context.Context) ServiceMiddleware {
	return func(next routing.Service) routing.Service {
		var e endpoint.Endpoint
		e = makeFetchRoutesEndpoint(ctx, proxyURL)
		e = circuitbreaker.Hystrix("fetch-routes")(e)
		return proxyRoutingService{ctx, e, next}
	}
}

func makeFetchRoutesEndpoint(ctx context.Context, instance string) endpoint.Endpoint {
	u, err := url.Parse(instance)
	if err != nil {
		panic(err)
	}
	if u.Path == "" {
		u.Path = "/paths"
	}
	return kithttp.NewClient(
		"GET", u,
		encodeFetchRoutesRequest,
		decodeFetchRoutesResponse,
	).Endpoint()
}
