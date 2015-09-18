package loadbalancer

import "github.com/go-kit/kit/endpoint"

// Factory is a function that converts an instance string, e.g. a host:port,
// to a usable endpoint. Factories are used by load balancers to convert
// instances returned by Publishers (typically host:port strings) into
// endpoints. Users are expected to provide their own factory functions that
// assume specific transports, or can deduce transports by parsing the
// instance string.
type Factory func(instance string) (endpoint.Endpoint, Closer, error)

// Closer is returned by factory functions as a way to close a generated
// endpoint.
type Closer chan struct{}
