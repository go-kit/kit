package loadbalancer_test

import (
	"runtime"
	"sync"

	"github.com/go-kit/kit/endpoint"
)

type mockPublisher struct {
	sync.Mutex
	e []endpoint.Endpoint
	s map[chan<- []endpoint.Endpoint]struct{}
}

func newMockPublisher(endpoints []endpoint.Endpoint) *mockPublisher {
	return &mockPublisher{
		e: endpoints,
		s: map[chan<- []endpoint.Endpoint]struct{}{},
	}
}

func (p *mockPublisher) replace(endpoints []endpoint.Endpoint) {
	p.Lock()
	defer p.Unlock()
	p.e = endpoints
	for s := range p.s {
		s <- p.e
	}
	runtime.Gosched()
}

func (p *mockPublisher) Subscribe(c chan<- []endpoint.Endpoint) {
	p.Lock()
	defer p.Unlock()
	p.s[c] = struct{}{}
	c <- p.e
}

func (p *mockPublisher) Unsubscribe(c chan<- []endpoint.Endpoint) {
	p.Lock()
	defer p.Unlock()
	delete(p.s, c)
}

func (p *mockPublisher) Stop() {}
