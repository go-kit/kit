package sd

import "github.com/go-kit/kit/endpoint"

// StaticService yields a fixed set of endpoints.
type StaticService map[string]endpoint.Endpoint

// Endpoint implements Service.
func (s StaticService) Endpoint(method string) (endpoint.Endpoint, error) {
	e, ok := s[method]
	if !ok {
		return nil, ErrMethodNotFound
	}
	return e, nil
}
