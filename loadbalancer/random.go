package loadbalancer

import (
	"math/rand"

	"github.com/go-kit/kit/endpoint"
)

// Random returns a load balancer that yields random endpoints.
func Random(p Publisher) LoadBalancer {
	return random{newCache(p)}
}

type random struct{ *cache }

func (r random) Count() int { return r.cache.count() }

func (r random) Get() (endpoint.Endpoint, error) {
	endpoints := r.cache.get()
	if len(endpoints) <= 0 {
		return nil, ErrNoEndpointsAvailable
	}
	return endpoints[rand.Intn(len(endpoints))], nil
}
