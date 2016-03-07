package handling

import (
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/voyage"
	"golang.org/x/net/context"
)

type registerIncidentRequest struct {
	ID             cargo.TrackingID
	Location       location.UNLocode
	Voyage         voyage.Number
	EventType      cargo.HandlingEventType
	CompletionTime time.Time
}

type registerIncidentResponse struct {
	Err error `json:"error,omitempty"`
}

func (r registerIncidentResponse) error() error { return r.Err }

func makeRegisterIncidentEndpoint(hs Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(registerIncidentRequest)
		err := hs.RegisterHandlingEvent(req.CompletionTime, req.ID, req.Voyage, req.Location, req.EventType)
		return registerIncidentResponse{Err: err}, nil
	}
}
