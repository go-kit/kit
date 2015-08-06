package loadbalancer

import (
	"sync/atomic"

	"github.com/go-kit/kit/endpoint"
)

// RoundRobin is a simple load balancer that returns each of the published
// endpoints in sequence.
type RoundRobin struct {
	p       Publisher
	counter uint64
}

// NewRoundRobin returns a new RoundRobin load balancer.
func NewRoundRobin(p Publisher) *RoundRobin {
	return &RoundRobin{
		p:       p,
		counter: 0,
	}
}

// Endpoint implements the LoadBalancer interface.
func (rr *RoundRobin) Endpoint() (endpoint.Endpoint, error) {
	endpoints, err := rr.p.Endpoints()
	if err != nil {
		return nil, err
	}
	if len(endpoints) <= 0 {
		return nil, ErrNoEndpoints
	}
	var old uint64
	for {
		old = atomic.LoadUint64(&rr.counter)
		if atomic.CompareAndSwapUint64(&rr.counter, old, old+1) {
			break
		}
	}
	return endpoints[old%uint64(len(endpoints))], nil
}
