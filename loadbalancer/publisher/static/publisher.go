package static

import (
	"sync"

	"github.com/go-kit/kit/endpoint"
)

// Publisher holds a static set of endpoints.
type Publisher struct {
	mu          sync.Mutex
	current     []endpoint.Endpoint
	subscribers map[chan<- []endpoint.Endpoint]struct{}
}

// NewPublisher returns a publisher that yields a static set of endpoints,
// which can be completely replaced.
func NewPublisher(endpoints []endpoint.Endpoint) *Publisher {
	return &Publisher{
		current:     endpoints,
		subscribers: map[chan<- []endpoint.Endpoint]struct{}{},
	}
}

// Subscribe implements Publisher.
func (p *Publisher) Subscribe(c chan<- []endpoint.Endpoint) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.subscribers[c] = struct{}{}
	c <- p.current
}

// Unsubscribe implements Publisher.
func (p *Publisher) Unsubscribe(c chan<- []endpoint.Endpoint) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.subscribers, c)
}

// Stop implements Publisher, but is a no-op.
func (p *Publisher) Stop() {}

// Replace replaces the endpoints and notifies all subscribers.
func (p *Publisher) Replace(endpoints []endpoint.Endpoint) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.current = endpoints
	for c := range p.subscribers {
		c <- p.current
	}
}
