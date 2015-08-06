package loadbalancer

import (
	"math/rand"

	"github.com/go-kit/kit/endpoint"
)

// Random is a completely stateless load balancer that chooses a random
// endpoint to return each time.
type Random struct {
	p Publisher
	r *rand.Rand
}

// NewRandom returns a new Random load balancer.
func NewRandom(p Publisher, seed int64) *Random {
	return &Random{
		p: p,
		r: rand.New(rand.NewSource(seed)),
	}
}

// Endpoint implements the LoadBalancer interface.
func (r *Random) Endpoint() (endpoint.Endpoint, error) {
	endpoints, err := r.p.Endpoints()
	if err != nil {
		return nil, err
	}
	if len(endpoints) <= 0 {
		return nil, ErrNoEndpoints
	}
	return endpoints[r.r.Intn(len(endpoints))], nil
}
