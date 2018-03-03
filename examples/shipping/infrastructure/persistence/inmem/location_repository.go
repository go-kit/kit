// Package inmem provides in-memory implementations of all the domain repositories.
package inmem

import (
	"github.com/go-kit/kit/examples/shipping/domain/model/location"
	"github.com/go-kit/kit/examples/shipping/domain/repository"
)

type locationRepository struct {
	locations map[location.UNLocode]*location.Location
}

func (r *locationRepository) Find(locode location.UNLocode) (*location.Location, error) {
	if l, ok := r.locations[locode]; ok {
		return l, nil
	}
	return nil, location.ErrUnknown
}

func (r *locationRepository) FindAll() []*location.Location {
	l := make([]*location.Location, 0, len(r.locations))
	for _, val := range r.locations {
		l = append(l, val)
	}
	return l
}

// NewLocationRepository returns a new instance of a in-memory location repository.
func NewLocationRepository() repository.LocationRepository {
	r := &locationRepository{
		locations: make(map[location.UNLocode]*location.Location),
	}

	r.locations[location.SESTO] = location.Stockholm
	r.locations[location.AUMEL] = location.Melbourne
	r.locations[location.CNHKG] = location.Hongkong
	r.locations[location.JNTKO] = location.Tokyo
	r.locations[location.NLRTM] = location.Rotterdam
	r.locations[location.DEHAM] = location.Hamburg

	return r
}
