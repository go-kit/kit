// Package inmem provides in-memory implementations of all the domain repositories.
package inmem

import (
	"sync"

	"github.com/go-kit/kit/examples/shipping/domain/model/cargo"
	"github.com/go-kit/kit/examples/shipping/domain/repository"
)

type handlingEventRepository struct {
	mtx    sync.RWMutex
	events map[cargo.TrackingID][]cargo.HandlingEvent
}

func (r *handlingEventRepository) Store(e cargo.HandlingEvent) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	// Make array if it's the first event with this tracking ID.
	if _, ok := r.events[e.TrackingID]; !ok {
		r.events[e.TrackingID] = make([]cargo.HandlingEvent, 0)
	}
	r.events[e.TrackingID] = append(r.events[e.TrackingID], e)
}

func (r *handlingEventRepository) QueryHandlingHistory(id cargo.TrackingID) cargo.HandlingHistory {
	r.mtx.RLock()
	defer r.mtx.RUnlock()
	return cargo.HandlingHistory{HandlingEvents: r.events[id]}
}

// NewHandlingEventRepository returns a new instance of a in-memory handling event repository.
func NewHandlingEventRepository() repository.HandlingEventRepository {
	return &handlingEventRepository{
		events: make(map[cargo.TrackingID][]cargo.HandlingEvent),
	}
}
