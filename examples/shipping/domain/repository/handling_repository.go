package repository

import (
	"github.com/go-kit/kit/examples/shipping/domain/model/cargo"
)

// HandlingEventRepository provides access a handling event store.
type HandlingEventRepository interface {
	Store(e cargo.HandlingEvent)
	QueryHandlingHistory(cargo.TrackingID) cargo.HandlingHistory
}
