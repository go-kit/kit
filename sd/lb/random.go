package lb

import (
	"math/rand"

	"github.com/go-kit/kit/sd"
)

// NewRandom returns a load balancer that selects services randomly.
func NewRandom(s sd.Subscriber, seed int64) Balancer {
	return &random{
		s: s,
		r: rand.New(rand.NewSource(seed)),
	}
}

type random struct {
	s sd.Subscriber
	r *rand.Rand
}

func (r *random) Service() (sd.Service, error) {
	services, err := r.s.Services()
	if err != nil {
		return nil, err
	}
	if len(services) <= 0 {
		return nil, ErrNoServices
	}
	return services[r.r.Intn(len(services))], nil
}
