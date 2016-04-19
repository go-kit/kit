package sd

import (
	"io"

	"github.com/go-kit/kit/service"
)

// Factory is a function that converts an instance string (e.g. host:port) to a
// service. It also returns an io.Closer that's invoked when the instance goes
// away and needs to be cleaned up. Users are expected to provide their own
// factory functions that assume specific transports, or can deduce transports
// by parsing the instance string.
type Factory func(instance string) (service.Service, io.Closer, error)
