package loadbalancer

import "errors"

// ErrNoEndpoints is returned when a load balancer (or one of its components)
// has no endpoints to return. In a request lifecycle, this is usually a fatal
// error.
var ErrNoEndpoints = errors.New("no endpoints available")
