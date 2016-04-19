package lb

import (
	"errors"

	"github.com/go-kit/kit/service"
)

// Balancer yields services according to some heuristic.
type Balancer interface {
	Service() (service.Service, error)
}

// ErrNoServices is returned when no qualifying services are available.
var ErrNoServices = errors.New("no services available")
