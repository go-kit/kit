package loadbalancer

import "github.com/go-kit/kit/endpoint"

// Publisher describes something that provides a set of identical endpoints.
// Different publisher implementations exist for different kinds of service
// discovery systems.
type Publisher interface {
	Endpoints() ([]endpoint.Endpoint, error)
}
