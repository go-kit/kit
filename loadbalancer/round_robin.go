package loadbalancer

import (
	"sync/atomic"

	"github.com/go-kit/kit/endpoint"
)

// RoundRobin returns a each endpoint from the most current set of endpoints
// in sequence.
func RoundRobin(p Publisher) Strategy {
	return &roundRobinStrategy{newEndpointCache(p), 0}
}

type roundRobinStrategy struct {
	cache  *endpointCache
	cursor uint64
}

func (s *roundRobinStrategy) Next() (endpoint.Endpoint, error) {
	endpoints := s.cache.get()
	if len(endpoints) <= 0 {
		return nil, ErrNoEndpoints
	}
	var cursor uint64
	for {
		cursor = atomic.LoadUint64(&s.cursor)
		if atomic.CompareAndSwapUint64(&s.cursor, cursor, cursor+1) {
			break
		}
	}
	return endpoints[cursor%uint64(len(endpoints))], nil
}

func (s *roundRobinStrategy) Stop() {
	s.cache.stop()
}
