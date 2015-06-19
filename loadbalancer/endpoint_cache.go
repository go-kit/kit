package loadbalancer

import "github.com/go-kit/kit/endpoint"

type endpointCache struct {
	requests chan []endpoint.Endpoint
	quit     chan struct{}
}

func newEndpointCache(p Publisher) *endpointCache {
	c := &endpointCache{
		requests: make(chan []endpoint.Endpoint),
		quit:     make(chan struct{}),
	}
	go c.loop(p)
	return c
}

func (c *endpointCache) loop(p Publisher) {
	updates := make(chan []endpoint.Endpoint, 1)
	p.Subscribe(updates)
	defer p.Unsubscribe(updates)
	endpoints := <-updates

	for {
		select {
		case endpoints = <-updates:
		case c.requests <- endpoints:
		case <-c.quit:
			return
		}
	}
}

func (c *endpointCache) get() []endpoint.Endpoint {
	return <-c.requests
}

func (c *endpointCache) stop() {
	close(c.quit)
}
