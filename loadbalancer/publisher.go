package loadbalancer

import (
	"errors"

	"github.com/go-kit/kit/endpoint"
)

// Publisher describes something that provides a set of identical endpoints.
// Different publisher implementations exist for different kinds of service
// discovery systems.
type Publisher interface {
	Endpoints() ([]endpoint.Endpoint, error)
}

// ErrPublisherStopped is returned by publishers when the underlying
// implementation has been terminated and can no longer serve requests.
var ErrPublisherStopped = errors.New("publisher stopped")
