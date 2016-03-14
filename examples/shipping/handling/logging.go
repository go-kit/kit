package handling

import (
	"time"

	"github.com/go-kit/kit/examples/shipping/cargo"
	"github.com/go-kit/kit/examples/shipping/location"
	"github.com/go-kit/kit/examples/shipping/voyage"
	"github.com/go-kit/kit/log"
)

type loggingService struct {
	logger log.Logger
	Service
}

// NewLoggingService returns a new instance of a logging Service.
func NewLoggingService(logger log.Logger, s Service) Service {
	return &loggingService{logger, s}
}

func (s *loggingService) RegisterHandlingEvent(completionTime time.Time, trackingID cargo.TrackingID, voyageNumber voyage.Number,
	unLocode location.UNLocode, eventType cargo.HandlingEventType) (err error) {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "register_incident",
			"tracking_id", trackingID,
			"location", unLocode,
			"voyage", voyageNumber,
			"event_type", eventType,
			"completion_time", completionTime,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return s.Service.RegisterHandlingEvent(completionTime, trackingID, voyageNumber, unLocode, eventType)
}
