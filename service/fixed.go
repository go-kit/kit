package service

import "github.com/go-kit/kit/endpoint"

// Fixed yields a fixed set of endpoints.
type Fixed map[string]endpoint.Endpoint

// Endpoint implements Service.
func (f Fixed) Endpoint(method string) (endpoint.Endpoint, error) {
	e, ok := f[method]
	if !ok {
		return nil, ErrMethodNotFound
	}
	return e, nil
}
