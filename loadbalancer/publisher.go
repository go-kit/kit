package loadbalancer

import "github.com/go-kit/kit/endpoint"

// Publisher publishes all available endpoints for a single service.
// Strategies (random, round-robin, least-used, etc.) can be built on top of
// publishers.
type Publisher interface {
	Subscribe(chan<- []endpoint.Endpoint)
	Unsubscribe(chan<- []endpoint.Endpoint)
	Stop()
}

// NewStaticPublisher returns a publisher that emits a fixed set of endpoints
// to every subscriber.
func NewStaticPublisher(endpoints []endpoint.Endpoint) Publisher {
	return staticPublisher(endpoints)
}

type staticPublisher []endpoint.Endpoint

func (p staticPublisher) Subscribe(c chan<- []endpoint.Endpoint) { c <- p }
func (p staticPublisher) Unsubscribe(chan<- []endpoint.Endpoint) {}
func (p staticPublisher) Stop()                                  {}
