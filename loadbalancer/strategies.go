package loadbalancer

import (
	"errors"

	"github.com/go-kit/kit/endpoint"
)

// Strategy yields endpoints to consumers according to some algorithm.
type Strategy interface {
	Next() (endpoint.Endpoint, error)
	Stop()
}

// ErrNoEndpoints is returned by a strategy when there are no endpoints
// available.
var ErrNoEndpoints = errors.New("no endpoints available")
