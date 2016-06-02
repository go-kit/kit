package lb

import (
	"sync/atomic"

	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/service"
)

// NewRoundRobin returns a load balancer that returns services in sequence.
func NewRoundRobin(s sd.Subscriber) Balancer {
	return &roundRobin{
		s: s,
		c: 0,
	}
}

type roundRobin struct {
	s sd.Subscriber
	c uint64
}

func (rr *roundRobin) Service() (service.Service, error) {
	services, err := rr.s.Services()
	if err != nil {
		return nil, err
	}
	if len(services) <= 0 {
		return nil, ErrNoServices
	}
	old := atomic.AddUint64(&rr.c, 1) - 1
	idx := old % uint64(len(services))
	return services[idx], nil
}
