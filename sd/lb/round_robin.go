package lb

import (
	"sync/atomic"

	"github.com/go-kit/kit/sd"
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

func (rr *roundRobin) Service() (sd.Service, error) {
	services, err := rr.s.Services()
	if err != nil {
		return nil, err
	}
	if len(services) <= 0 {
		return nil, ErrNoServices
	}
	var old uint64
	for {
		old = atomic.LoadUint64(&rr.c)
		if atomic.CompareAndSwapUint64(&rr.c, old, old+1) {
			break
		}
	}
	return services[old%uint64(len(services))], nil
}
