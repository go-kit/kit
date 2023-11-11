package lb

import (
	"math/rand"

	"github.com/openmesh/kit/endpoint"
	"github.com/openmesh/kit/sd"
)

// NewRandom returns a load balancer that selects services randomly.
func NewRandom[Request, Response any](s sd.Endpointer[Request, Response], seed int64) Balancer[Request, Response] {
	return &random[Request, Response]{
		s: s,
		r: rand.New(rand.NewSource(seed)),
	}
}

type random[Request, Response any] struct {
	s sd.Endpointer[Request, Response]
	r *rand.Rand
}

func (r *random[Request, Response]) Endpoint() (endpoint.Endpoint[Request, Response], error) {
	endpoints, err := r.s.Endpoints()
	if err != nil {
		return nil, err
	}
	if len(endpoints) <= 0 {
		return nil, ErrNoEndpoints
	}
	return endpoints[r.r.Intn(len(endpoints))], nil
}
