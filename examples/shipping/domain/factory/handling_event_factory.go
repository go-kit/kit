package factory

import (
	"time"

	"github.com/go-kit/kit/examples/shipping/domain/model/cargo"
	"github.com/go-kit/kit/examples/shipping/domain/model/location"
	"github.com/go-kit/kit/examples/shipping/domain/model/voyage"
	"github.com/go-kit/kit/examples/shipping/domain/repository"
)

// HandlingEventFactory creates handling events.
type HandlingEventFactory struct {
	CargoRepository    repository.CargoRepository
	VoyageRepository   repository.VoyageRepository
	LocationRepository repository.LocationRepository
}

// CreateHandlingEvent creates a validated handling event.
func (f *HandlingEventFactory) CreateHandlingEvent(registered time.Time, completed time.Time, id cargo.TrackingID,
	voyageNumber voyage.Number, unLocode location.UNLocode, eventType cargo.HandlingEventType) (cargo.HandlingEvent, error) {

	if _, err := f.CargoRepository.Find(id); err != nil {
		return cargo.HandlingEvent{}, err
	}

	if _, err := f.VoyageRepository.Find(voyageNumber); err != nil {
		// TODO: This is pretty ugly, but when creating a Receive event, the voyage number is not known.
		if len(voyageNumber) > 0 {
			return cargo.HandlingEvent{}, err
		}
	}

	if _, err := f.LocationRepository.Find(unLocode); err != nil {
		return cargo.HandlingEvent{}, err
	}

	return cargo.HandlingEvent{
		TrackingID: id,
		Activity: cargo.HandlingActivity{
			Type:         eventType,
			Location:     unLocode,
			VoyageNumber: voyageNumber,
		},
	}, nil
}
