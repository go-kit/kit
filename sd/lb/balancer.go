package lb

import (
	"errors"

	"github.com/go-kit/kit/sd"
)

// Balancer yields services according to some heuristic.
type Balancer interface {
	Service() (sd.Service, error)
}

// ErrNoServices is returned when no qualifying services are available.
var ErrNoServices = errors.New("no services available")
