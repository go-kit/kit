package loadbalancer

import (
	"io"
	"sync"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

// EndpointCache caches endpoints that need to be deallocated when they're no
// longer useful. Clients update the cache by providing a current set of
// instance strings. The cache converts each instance string to an endpoint
// and a closer via the factory function.
//
// Instance strings are assumed to be unique and are used as keys. Endpoints
// that were in the previous set of instances and are not in the current set
// are considered invalid and closed.
//
// EndpointCache is designed to be used in your publisher implementation.
type EndpointCache struct {
	mtx    sync.RWMutex
	f      Factory
	m      map[string]endpointCloser
	logger log.Logger
}

// NewEndpointCache produces a new EndpointCache, ready for use. Instance
// strings will be converted to endpoints via the provided factory function.
// The logger is used to log errors.
func NewEndpointCache(f Factory, logger log.Logger) *EndpointCache {
	return &EndpointCache{
		f:      f,
		m:      map[string]endpointCloser{},
		logger: log.NewContext(logger).With("component", "Endpoint Cache"),
	}
}

type endpointCloser struct {
	endpoint.Endpoint
	io.Closer
}

// Replace replaces the current set of endpoints with endpoints manufactured
// by the passed instances. If the same instance exists in both the existing
// and new sets, it's left untouched.
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
		if ec.Closer != nil {
			ec.Closer.Close()
		}
	}

	// Swap and GC.
	t.m = m
}

// Endpoints returns the current set of endpoints in undefined order. Satisfies
// Publisher interface.
func (t *EndpointCache) Endpoints() ([]endpoint.Endpoint, error) {
	t.mtx.RLock()
	defer t.mtx.RUnlock()
	a := make([]endpoint.Endpoint, 0, len(t.m))
	for _, ec := range t.m {
		a = append(a, ec.Endpoint)
	}
	return a, nil
}
