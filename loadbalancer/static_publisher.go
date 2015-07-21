package loadbalancer

import (
	"sync"

	"github.com/go-kit/kit/endpoint"
)

// NewStaticPublisher returns a publisher that yields a static set of
// endpoints, which can be completely replaced.
func NewStaticPublisher(endpoints []endpoint.Endpoint) *StaticPublisher {
	return &StaticPublisher{
		current:     endpoints,
		subscribers: map[chan<- []endpoint.Endpoint]struct{}{},
	}
}

// StaticPublisher holds a static set of endpoints.
type StaticPublisher struct {
	mu          sync.Mutex
	current     []endpoint.Endpoint
	subscribers map[chan<- []endpoint.Endpoint]struct{}
}

// Subscribe implements Publisher.
func (p *StaticPublisher) Subscribe(c chan<- []endpoint.Endpoint) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.subscribers[c] = struct{}{}
	c <- p.current
}

// Unsubscribe implements Publisher.
func (p *StaticPublisher) Unsubscribe(c chan<- []endpoint.Endpoint) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.subscribers, c)
}

// Stop implements Publisher, but is a no-op.
func (p *StaticPublisher) Stop() {}

// Replace replaces the endpoints and notifies all subscribers.
func (p *StaticPublisher) Replace(endpoints []endpoint.Endpoint) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.current = endpoints
	for c := range p.subscribers {
		c <- p.current
	}
}
