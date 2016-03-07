// Package booking provides the use-case of booking a cargo. Used by views
// facing an administrator.
package booking

import (
	"errors"
	"time"

	"github.com/go-kit/kit/examples/shipping/cargo"
	"github.com/go-kit/kit/examples/shipping/location"
	"github.com/go-kit/kit/examples/shipping/routing"
)

// ErrInvalidArgument is returned when one or more arguments are invalid.
var ErrInvalidArgument = errors.New("invalid argument")

// Service is the interface that provides booking methods.
type Service interface {
	// BookNewCargo registers a new cargo in the tracking system, not yet
	// routed.
	BookNewCargo(origin location.UNLocode, destination location.UNLocode, arrivalDeadline time.Time) (cargo.TrackingID, error)

	// LoadCargo returns a read model of a cargo.
	LoadCargo(trackingID cargo.TrackingID) (Cargo, error)

	// RequestPossibleRoutesForCargo requests a list of itineraries describing
	// possible routes for this cargo.
	RequestPossibleRoutesForCargo(trackingID cargo.TrackingID) []cargo.Itinerary

	// AssignCargoToRoute assigns a cargo to the route specified by the
	// itinerary.
	AssignCargoToRoute(trackingID cargo.TrackingID, itinerary cargo.Itinerary) error

	// ChangeDestination changes the destination of a cargo.
	ChangeDestination(trackingID cargo.TrackingID, unLocode location.UNLocode) error

	// Cargos returns a list of all cargos that have been booked.
	Cargos() []Cargo

	// Locations returns a list of registered locations.
	Locations() []Location
}

type service struct {
	cargoRepository         cargo.Repository
	locationRepository      location.Repository
	routingService          routing.Service
	handlingEventRepository cargo.HandlingEventRepository
}

func (s *service) AssignCargoToRoute(id cargo.TrackingID, itinerary cargo.Itinerary) error {
	if id == "" || len(itinerary.Legs) == 0 {
		return ErrInvalidArgument
	}

	c, err := s.cargoRepository.Find(id)
	if err != nil {
		return err
	}

	c.AssignToRoute(itinerary)

	if err := s.cargoRepository.Store(c); err != nil {
		return err
	}

	return nil
}

func (s *service) BookNewCargo(origin, destination location.UNLocode, arrivalDeadline time.Time) (cargo.TrackingID, error) {
	if origin == "" || destination == "" || arrivalDeadline.IsZero() {
		return "", ErrInvalidArgument
	}

	id := cargo.NextTrackingID()
	rs := cargo.RouteSpecification{
		Origin:          origin,
		Destination:     destination,
		ArrivalDeadline: arrivalDeadline,
	}

	c := cargo.New(id, rs)

	if err := s.cargoRepository.Store(c); err != nil {
		return "", err
	}

	return c.TrackingID, nil
}

func (s *service) LoadCargo(trackingID cargo.TrackingID) (Cargo, error) {
	if trackingID == "" {
		return Cargo{}, ErrInvalidArgument
	}

	c, err := s.cargoRepository.Find(trackingID)
	if err != nil {
		return Cargo{}, err
	}

	return assemble(c, s.handlingEventRepository), nil
}

func (s *service) ChangeDestination(id cargo.TrackingID, destination location.UNLocode) error {
	if id == "" || destination == "" {
		return ErrInvalidArgument
	}

	c, err := s.cargoRepository.Find(id)
	if err != nil {
		return err
	}

	l, err := s.locationRepository.Find(destination)
	if err != nil {
		return err
	}

	c.SpecifyNewRoute(cargo.RouteSpecification{
		Origin:          c.Origin,
		Destination:     l.UNLocode,
		ArrivalDeadline: c.RouteSpecification.ArrivalDeadline,
	})

	if err := s.cargoRepository.Store(c); err != nil {
		return err
	}

	return nil
}

func (s *service) RequestPossibleRoutesForCargo(id cargo.TrackingID) []cargo.Itinerary {
	if id == "" {
		return nil
	}

	c, err := s.cargoRepository.Find(id)
	if err != nil {
		return []cargo.Itinerary{}
	}

	return s.routingService.FetchRoutesForSpecification(c.RouteSpecification)
}

func (s *service) Cargos() []Cargo {
	var result []Cargo
	for _, c := range s.cargoRepository.FindAll() {
		result = append(result, assemble(c, s.handlingEventRepository))
	}
	return result
}

func (s *service) Locations() []Location {
	var result []Location
	for _, v := range s.locationRepository.FindAll() {
		result = append(result, Location{
			UNLocode: string(v.UNLocode),
			Name:     v.Name,
		})
	}
	return result
}

// NewService creates a booking service with necessary dependencies.
func NewService(cr cargo.Repository, lr location.Repository, her cargo.HandlingEventRepository, rs routing.Service) Service {
	return &service{
		cargoRepository:         cr,
		locationRepository:      lr,
		handlingEventRepository: her,
		routingService:          rs,
	}
}

// Location is a read model for booking views.
type Location struct {
	UNLocode string `json:"locode"`
	Name     string `json:"name"`
}

// Cargo is a read model for booking views.
type Cargo struct {
	ArrivalDeadline time.Time   `json:"arrival_deadline"`
	Destination     string      `json:"destination"`
	Legs            []cargo.Leg `json:"legs,omitempty"`
	Misrouted       bool        `json:"misrouted"`
	Origin          string      `json:"origin"`
	Routed          bool        `json:"routed"`
	TrackingID      string      `json:"tracking_id"`
}

func assemble(c *cargo.Cargo, her cargo.HandlingEventRepository) Cargo {
	return Cargo{
		TrackingID:      string(c.TrackingID),
		Origin:          string(c.Origin),
		Destination:     string(c.RouteSpecification.Destination),
		Misrouted:       c.Delivery.RoutingStatus == cargo.Misrouted,
		Routed:          !c.Itinerary.IsEmpty(),
		ArrivalDeadline: c.RouteSpecification.ArrivalDeadline,
		Legs:            c.Itinerary.Legs,
	}
}
