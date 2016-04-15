package sd

import (
	"errors"

	"github.com/go-kit/kit/endpoint"
)

// Service represents a collection of endpoints (i.e. methods). It may be one
// instance of a microservice, or it may abstract over multiple identical
// instances of the same microservice.
type Service interface {
	Endpoint(method string) (endpoint.Endpoint, error)
}

// ErrMethodNotFound indicates a method has no corresponding endpoint.
var ErrMethodNotFound = errors.New("method not found")
