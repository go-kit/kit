package sd

import (
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

// Endpointer listens to a service discovery system and yields a set of
// identical endpoints on demand. An error indicates a problem with connectivity
// to the service discovery system, or within the system itself; an Endpointer
// may yield no endpoints without error.
type Endpointer interface {
	Endpoints() ([]endpoint.Endpoint, error)
}

// FixedEndpointer yields a fixed set of endpoints.
type FixedEndpointer []endpoint.Endpoint

// Endpoints implements Endpointer.
func (s FixedEndpointer) Endpoints() ([]endpoint.Endpoint, error) { return s, nil }

// NewEndpointer creates an Endpointer that subscribes to updates from Instancer src
// and uses factory f to create Endpoints. If src notifies of an error, the Endpointer
// keeps returning previously created Endpoints assuming they are still good, unless
// this behavior is disabled with ResetOnError option.
func NewEndpointer(src Instancer, f Factory, logger log.Logger, options ...EndpointerOption) Endpointer {
	opts := endpointerOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	se := &simpleEndpointer{
		endpointCache: *newEndpointCache(f, logger, opts),
		instancer:     src,
		ch:            make(chan Event),
	}
	go se.receive()
	src.Register(se.ch)
	return se
}

// EndpointerOption allows control of endpointCache behavior.
type EndpointerOption func(*endpointerOptions)

// InvalidateOnError returns EndpointerOption that controls how the Endpointer
// behaves when then Instancer publishes an Event containing an error.
// Without this option the Endpointer continues returning the last known
// endpoints. With this option, the Endpointer continues returning the last
// known endpoints until the timeout elapses, then closes all active endpoints
// and starts returning an error. Once the Instancer sends a new update with
// valid resource instances, the normal operation is resumed.
func InvalidateOnError(timeout time.Duration) EndpointerOption {
	return func(opts *endpointerOptions) {
		opts.invalidateOnErrorTimeout = &timeout
	}
}

type endpointerOptions struct {
	invalidateOnErrorTimeout *time.Duration
}

type simpleEndpointer struct {
	endpointCache

	instancer Instancer
	ch        chan Event
}

func (se *simpleEndpointer) receive() {
	for event := range se.ch {
		se.Update(event)
	}
}

func (se *simpleEndpointer) Close() {
	se.instancer.Deregister(se.ch)
	close(se.ch)
}
