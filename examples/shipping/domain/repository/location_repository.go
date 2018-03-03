package repository

import (
	"github.com/go-kit/kit/examples/shipping/domain/model/location"
)

// LocationRepository provides access a location store.
type LocationRepository interface {
	Find(locode location.UNLocode) (*location.Location, error)
	FindAll() []*location.Location
}
