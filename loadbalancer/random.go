package loadbalancer

import (
	"math/rand"

	"github.com/go-kit/kit/endpoint"
)

// Random returns a random endpoint from the most current set of endpoints.
func Random(p Publisher) Strategy {
	return &randomStrategy{newEndpointCache(p)}
}

type randomStrategy struct{ *endpointCache }

func (s randomStrategy) Next() (endpoint.Endpoint, error) {
	endpoints := s.endpointCache.get()
	if len(endpoints) <= 0 {
		return nil, ErrNoEndpoints
	}
	return endpoints[rand.Intn(len(endpoints))], nil
}

func (s randomStrategy) Stop() {
	s.endpointCache.stop()
}
