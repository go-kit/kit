package loadbalancer

import (
	"sync"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

// EndpointCache TODO
type EndpointCache struct {
	mtx    sync.RWMutex
	f      Factory
	m      map[string]endpointCloser
	logger log.Logger
}

// NewEndpointCache TODO
func NewEndpointCache(f Factory, logger log.Logger) *EndpointCache {
	return &EndpointCache{
		f:      f,
		m:      map[string]endpointCloser{},
		logger: logger,
	}
}

type endpointCloser struct {
	endpoint.Endpoint
	Closer
}

// Replace TODO
func (t *EndpointCache) Replace(instances []string) {
	t.mtx.Lock()
	defer t.mtx.Unlock()

	// Produce the current set of endpoints.
	m := make(map[string]endpointCloser, len(instances))
	for _, instance := range instances {
		// If it already exists, just copy it over.
		if ec, ok := t.m[instance]; ok {
			m[instance] = ec
			delete(t.m, instance)
			continue
		}

		// If it doesn't exist, create it.
		endpoint, closer, err := t.f(instance)
		if err != nil {
			t.logger.Log("instance", instance, "err", err)
			continue
		}
		m[instance] = endpointCloser{endpoint, closer}
	}

	// Close any leftover endpoints.
	for _, ec := range t.m {
		close(ec.Closer)
	}

	// Swap and GC.
	t.m = m
}

// Endpoints TODO
func (t *EndpointCache) Endpoints() []endpoint.Endpoint {
	t.mtx.RLock()
	defer t.mtx.RUnlock()
	a := make([]endpoint.Endpoint, 0, len(t.m))
	for _, ec := range t.m {
		a = append(a, ec.Endpoint)
	}
	return a
}
