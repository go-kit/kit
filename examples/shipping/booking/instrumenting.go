package booking

import (
	"time"

	"github.com/go-kit/kit/metrics"

	"github.com/go-kit/kit/examples/shipping/cargo"
	"github.com/go-kit/kit/examples/shipping/location"
)

type instrumentingService struct {
	requestCount   metrics.Counter
	requestLatency metrics.TimeHistogram
	Service
}

// NewInstrumentingService returns an instance of an instrumenting Service.
func NewInstrumentingService(requestCount metrics.Counter, requestLatency metrics.TimeHistogram, s Service) Service {
	return &instrumentingService{
		requestCount:   requestCount,
		requestLatency: requestLatency,
		Service:        s,
	}
}

func (s *instrumentingService) BookNewCargo(origin, destination location.UNLocode, arrivalDeadline time.Time) (cargo.TrackingID, error) {
	defer func(begin time.Time) {
		methodField := metrics.Field{Key: "method", Value: "book"}
		s.requestCount.With(methodField).Add(1)
		s.requestLatency.With(methodField).Observe(time.Since(begin))
	}(time.Now())

	return s.Service.BookNewCargo(origin, destination, arrivalDeadline)
}

func (s *instrumentingService) LoadCargo(id cargo.TrackingID) (c Cargo, err error) {
	defer func(begin time.Time) {
		methodField := metrics.Field{Key: "method", Value: "load"}
		s.requestCount.With(methodField).Add(1)
		s.requestLatency.With(methodField).Observe(time.Since(begin))
	}(time.Now())

	return s.Service.LoadCargo(id)
}

func (s *instrumentingService) RequestPossibleRoutesForCargo(id cargo.TrackingID) []cargo.Itinerary {
	defer func(begin time.Time) {
		methodField := metrics.Field{Key: "method", Value: "request_routes"}
		s.requestCount.With(methodField).Add(1)
		s.requestLatency.With(methodField).Observe(time.Since(begin))
	}(time.Now())

	return s.Service.RequestPossibleRoutesForCargo(id)
}

func (s *instrumentingService) AssignCargoToRoute(id cargo.TrackingID, itinerary cargo.Itinerary) (err error) {
	defer func(begin time.Time) {
		methodField := metrics.Field{Key: "method", Value: "assign_to_route"}
		s.requestCount.With(methodField).Add(1)
		s.requestLatency.With(methodField).Observe(time.Since(begin))
	}(time.Now())

	return s.Service.AssignCargoToRoute(id, itinerary)
}

func (s *instrumentingService) ChangeDestination(id cargo.TrackingID, l location.UNLocode) (err error) {
	defer func(begin time.Time) {
		methodField := metrics.Field{Key: "method", Value: "change_destination"}
		s.requestCount.With(methodField).Add(1)
		s.requestLatency.With(methodField).Observe(time.Since(begin))
	}(time.Now())

	return s.Service.ChangeDestination(id, l)
}

func (s *instrumentingService) Cargos() []Cargo {
	defer func(begin time.Time) {
		methodField := metrics.Field{Key: "method", Value: "list_cargos"}
		s.requestCount.With(methodField).Add(1)
		s.requestLatency.With(methodField).Observe(time.Since(begin))
	}(time.Now())

	return s.Service.Cargos()
}

func (s *instrumentingService) Locations() []Location {
	defer func(begin time.Time) {
		methodField := metrics.Field{Key: "method", Value: "list_locations"}
		s.requestCount.With(methodField).Add(1)
		s.requestLatency.With(methodField).Observe(time.Since(begin))
	}(time.Now())

	return s.Service.Locations()
}
