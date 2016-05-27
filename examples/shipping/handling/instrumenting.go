package handling

import (
	"time"

	"github.com/go-kit/kit/metrics"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"

	"github.com/go-kit/kit/examples/shipping/cargo"
	"github.com/go-kit/kit/examples/shipping/location"
	"github.com/go-kit/kit/examples/shipping/voyage"
)

type instrumentingService struct {
	requestCount   metrics.Counter
	requestLatency metrics.TimeHistogram
	Service
}

// NewInstrumentingService returns an instance of an instrumenting Service.
func NewInstrumentingService(s Service) Service {
	fieldKeys := []string{"method"}

	requestCount := kitprometheus.NewCounter(stdprometheus.CounterOpts{
		Namespace: "api",
		Subsystem: "handling_service",
		Name:      "request_count",
		Help:      "Number of requests received.",
	}, fieldKeys)

	requestLatency := metrics.NewTimeHistogram(time.Microsecond, kitprometheus.NewSummary(stdprometheus.SummaryOpts{
		Namespace: "api",
		Subsystem: "handling_service",
		Name:      "request_latency_microseconds",
		Help:      "Total duration of requests in microseconds.",
	}, fieldKeys))

	return &instrumentingService{
		requestCount:   requestCount,
		requestLatency: requestLatency,
		Service:        s,
	}
}

func (s *instrumentingService) RegisterHandlingEvent(completionTime time.Time, trackingID cargo.TrackingID, voyage voyage.Number,
	loc location.UNLocode, eventType cargo.HandlingEventType) error {

	defer func(begin time.Time) {
		methodField := metrics.Field{Key: "method", Value: "register_incident"}
		s.requestCount.With(methodField).Add(1)
		s.requestLatency.With(methodField).Observe(time.Since(begin))
	}(time.Now())

	return s.Service.RegisterHandlingEvent(completionTime, trackingID, voyage, loc, eventType)
}
