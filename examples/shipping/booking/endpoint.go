package booking

import (
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/routing"
	"golang.org/x/net/context"
)

type bookCargoRequest struct {
	Origin          location.UNLocode
	Destination     location.UNLocode
	ArrivalDeadline time.Time
}

type bookCargoResponse struct {
	ID  cargo.TrackingID `json:"tracking_id,omitempty"`
	Err error            `json:"error,omitempty"`
}

func (r bookCargoResponse) error() error { return r.Err }

func makeBookCargoEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(bookCargoRequest)
		id, err := s.BookNewCargo(req.Origin, req.Destination, req.ArrivalDeadline)
		return bookCargoResponse{ID: id, Err: err}, nil
	}
}

type loadCargoRequest struct {
	ID cargo.TrackingID
}

type loadCargoResponse struct {
	Cargo *Cargo `json:"cargo,omitempty"`
	Err   error  `json:"error,omitempty"`
}

func (r loadCargoResponse) error() error { return r.Err }

func makeLoadCargoEndpoint(bs Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(loadCargoRequest)
		c, err := bs.LoadCargo(req.ID)
		return loadCargoResponse{Cargo: &c, Err: err}, nil
	}
}

type requestRoutesRequest struct {
	ID cargo.TrackingID
}

type requestRoutesResponse struct {
	Routes []routing.Route `json:"routes,omitempty"`
	Err    error           `json:"error,omitempty"`
}

func (r requestRoutesResponse) error() error { return r.Err }

func makeRequestRoutesEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(requestRoutesRequest)

		itineraries := s.RequestPossibleRoutesForCargo(req.ID)

		result := []routing.Route{}
		for _, itin := range itineraries {
			var legs []routing.Leg
			for _, leg := range itin.Legs {
				legs = append(legs, routing.Leg{
					VoyageNumber: string(leg.VoyageNumber),
					From:         string(leg.LoadLocation),
					To:           string(leg.UnloadLocation),
					LoadTime:     leg.LoadTime,
					UnloadTime:   leg.UnloadTime,
				})
			}
			result = append(result, routing.Route{Legs: legs})
		}

		return requestRoutesResponse{Routes: result, Err: nil}, nil
	}
}

type assignToRouteRequest struct {
	ID        cargo.TrackingID
	Itinerary cargo.Itinerary
}

type assignToRouteResponse struct {
	Err error `json:"error,omitempty"`
}

func (r assignToRouteResponse) error() error { return r.Err }

func makeAssignToRouteEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(assignToRouteRequest)
		err := s.AssignCargoToRoute(req.ID, req.Itinerary)
		return assignToRouteResponse{Err: err}, nil
	}
}

type changeDestinationRequest struct {
	ID          cargo.TrackingID
	Destination location.UNLocode
}

type changeDestinationResponse struct {
	Err error `json:"error,omitempty"`
}

func (r changeDestinationResponse) error() error { return r.Err }

func makeChangeDestinationEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(changeDestinationRequest)
		err := s.ChangeDestination(req.ID, req.Destination)
		return changeDestinationResponse{Err: err}, nil
	}
}

type listCargosRequest struct{}

type listCargosResponse struct {
	Cargos []Cargo `json:"cargos,omitempty"`
	Err    error   `json:"error,omitempty"`
}

func (r listCargosResponse) error() error { return r.Err }

func makeListCargosEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		_ = request.(listCargosRequest)
		return listCargosResponse{Cargos: s.Cargos(), Err: nil}, nil
	}
}

type listLocationsRequest struct {
}

type listLocationsResponse struct {
	Locations []Location `json:"locations,omitempty"`
	Err       error      `json:"error,omitempty"`
}

func makeListLocationsEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		_ = request.(listLocationsRequest)
		return listLocationsResponse{Locations: s.Locations(), Err: nil}, nil
	}
}

type fetchRoutesRequest struct {
	From string
	To   string
}

type fetchRoutesResponse struct {
	Paths []struct {
		Edges []struct {
			Origin      string    `json:"origin"`
			Destination string    `json:"destination"`
			Voyage      string    `json:"voyage"`
			Departure   time.Time `json:"departure"`
			Arrival     time.Time `json:"arrival"`
		} `json:"edges"`
	} `json:"paths"`
}
