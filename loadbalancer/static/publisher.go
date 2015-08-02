package static

import "github.com/go-kit/kit/endpoint"

// Publisher yields the same set of static endpoints.
type Publisher []endpoint.Endpoint

// Endpoints implements the Publisher interface.
func (p Publisher) Endpoints() ([]endpoint.Endpoint, error) { return p, nil }
