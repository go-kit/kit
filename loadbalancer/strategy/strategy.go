package strategy

import "github.com/go-kit/kit/endpoint"

// Strategy yields endpoints to consumers according to some algorithm.
type Strategy interface {
	Next() (endpoint.Endpoint, error)
	Stop()
}
