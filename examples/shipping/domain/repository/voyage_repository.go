package repository

import (
	"github.com/go-kit/kit/examples/shipping/domain/model/voyage"
)

// VoyageRepository provides access a voyage store.
type VoyageRepository interface {
	Find(voyage.Number) (*voyage.Voyage, error)
}
