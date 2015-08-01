package strategy

import (
	"sync/atomic"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/loadbalancer"
	"github.com/go-kit/kit/loadbalancer/publisher"
)

// RoundRobin returns a load balancer that yields endpoints in sequence.
func RoundRobin(p publisher.Publisher) loadbalancer.LoadBalancer {
	return &roundRobin{newCache(p), 0}
}

type roundRobin struct {
	*cache
	uint64
}

func (r *roundRobin) Count() int { return r.cache.count() }

func (r *roundRobin) Get() (endpoint.Endpoint, error) {
	endpoints := r.cache.get()
	if len(endpoints) <= 0 {
		return nil, loadbalancer.ErrNoEndpointsAvailable
	}
	var old uint64
	for {
		old = atomic.LoadUint64(&r.uint64)
		if atomic.CompareAndSwapUint64(&r.uint64, old, old+1) {
			break
		}
	}
	return endpoints[old%uint64(len(endpoints))], nil
}
