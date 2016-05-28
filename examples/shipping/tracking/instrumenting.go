package tracking

import (
	"time"

	"github.com/go-kit/kit/metrics"
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

func (s *instrumentingService) Track(id string) (Cargo, error) {
	defer func(begin time.Time) {
		methodField := metrics.Field{Key: "method", Value: "track"}
		s.requestCount.With(methodField).Add(1)
		s.requestLatency.With(methodField).Observe(time.Since(begin))
	}(time.Now())

	return s.Service.Track(id)
}
