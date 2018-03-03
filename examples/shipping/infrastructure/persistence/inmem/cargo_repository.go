// Package inmem provides in-memory implementations of all the domain repositories.
package inmem

import (
	"sync"

	"github.com/go-kit/kit/examples/shipping/domain/model/cargo"
	"github.com/go-kit/kit/examples/shipping/domain/repository"
)

type cargoRepository struct {
	mtx    sync.RWMutex
	cargos map[cargo.TrackingID]*cargo.Cargo
}

func (r *cargoRepository) Store(c *cargo.Cargo) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	r.cargos[c.TrackingID] = c
	return nil
}

func (r *cargoRepository) Find(id cargo.TrackingID) (*cargo.Cargo, error) {
	r.mtx.RLock()
	defer r.mtx.RUnlock()
	if val, ok := r.cargos[id]; ok {
		return val, nil
	}
	return nil, cargo.ErrUnknown
}

func (r *cargoRepository) FindAll() []*cargo.Cargo {
	r.mtx.RLock()
	defer r.mtx.RUnlock()
	c := make([]*cargo.Cargo, 0, len(r.cargos))
	for _, val := range r.cargos {
		c = append(c, val)
	}
	return c
}

// NewCargoRepository returns a new instance of a in-memory cargo repository.
func NewCargoRepository() repository.CargoRepository {
	return &cargoRepository{
		cargos: make(map[cargo.TrackingID]*cargo.Cargo),
	}
}
