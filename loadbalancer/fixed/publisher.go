package fixed

import (
	"sync"

	"github.com/go-kit/kit/endpoint"
)

// Publisher yields the same set of fixed endpoints.
type Publisher struct {
	mtx       sync.RWMutex
	endpoints []endpoint.Endpoint
}

// NewPublisher returns a fixed endpoint Publisher.
func NewPublisher(endpoints []endpoint.Endpoint) *Publisher {
	return &Publisher{
		endpoints: endpoints,
	}
}

// Endpoints implements the Publisher interface.
func (p *Publisher) Endpoints() ([]endpoint.Endpoint, error) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()
	return p.endpoints, nil
}

// Replace is a utility method to swap out the underlying endpoints of an
// existing fixed publisher. It's useful mostly for testing.
func (p *Publisher) Replace(endpoints []endpoint.Endpoint) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.endpoints = endpoints
}
