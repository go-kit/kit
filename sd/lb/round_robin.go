package lb

import (
	"sync/atomic"

	"github.com/openmesh/kit/endpoint"
	"github.com/openmesh/kit/sd"
)

// NewRoundRobin returns a load balancer that returns services in sequence.
func NewRoundRobin[Request, Response any](s sd.Endpointer[Request, Response]) Balancer[Request, Response] {
	return &roundRobin[Request, Response]{
		s: s,
		c: 0,
	}
}

type roundRobin[Request, Response any] struct {
	s sd.Endpointer[Request, Response]
	c uint64
}

func (rr *roundRobin[Request, Response]) Endpoint() (endpoint.Endpoint[Request, Response], error) {
	endpoints, err := rr.s.Endpoints()
	if err != nil {
		return nil, err
	}
	if len(endpoints) <= 0 {
		return nil, ErrNoEndpoints
	}
	old := atomic.AddUint64(&rr.c, 1) - 1
	idx := old % uint64(len(endpoints))
	return endpoints[idx], nil
}
