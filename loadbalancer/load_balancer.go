package loadbalancer

import (
	"errors"

	"github.com/go-kit/kit/endpoint"
)

// LoadBalancer yields endpoints one-by-one.
type LoadBalancer interface {
	Count() int
	Get() (endpoint.Endpoint, error)
}

// ErrNoEndpointsAvailable is given by a load balancer when no endpoints are
// available to be returned.
var ErrNoEndpointsAvailable = errors.New("no endpoints available")
