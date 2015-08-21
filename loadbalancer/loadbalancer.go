package loadbalancer

import (
	"errors"

	"gopkg.in/kit.v0/endpoint"
)

// LoadBalancer describes something that can yield endpoints for a remote
// service method.
type LoadBalancer interface {
	Endpoint() (endpoint.Endpoint, error)
}

// ErrNoEndpoints is returned when a load balancer (or one of its components)
// has no endpoints to return. In a request lifecycle, this is usually a fatal
// error.
var ErrNoEndpoints = errors.New("no endpoints available")
