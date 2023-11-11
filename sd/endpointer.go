package sd

import (
	"time"

	"github.com/go-kit/log"
	"github.com/openmesh/kit/endpoint"
)

// Endpointer listens to a service discovery system and yields a set of
// identical endpoints on demand. An error indicates a problem with connectivity
// to the service discovery system, or within the system itself; an Endpointer
// may yield no endpoints without error.
type Endpointer[Request, Response any] interface {
	Endpoints() ([]endpoint.Endpoint[Request, Response], error)
}

// FixedEndpointer yields a fixed set of endpoints.
type FixedEndpointer[Request, Response any] []endpoint.Endpoint[Request, Response]

// Endpoints implements Endpointer.
func (s FixedEndpointer[Request, Response]) Endpoints() ([]endpoint.Endpoint[Request, Response], error) {
	return s, nil
}

// NewEndpointer creates an Endpointer that subscribes to updates from Instancer src
// and uses factory f to create Endpoints. If src notifies of an error, the Endpointer
// keeps returning previously created Endpoints assuming they are still good, unless
// this behavior is disabled via InvalidateOnError option.
func NewEndpointer[Request, Response any](src Instancer, f Factory[Request, Response], logger log.Logger, options ...EndpointerOption) *DefaultEndpointer[Request, Response] {
	opts := endpointerOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	se := &DefaultEndpointer[Request, Response]{
		cache:     newEndpointCache(f, logger, opts),
		instancer: src,
		ch:        make(chan Event),
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
		opts.invalidateOnError = true
		opts.invalidateTimeout = timeout
	}
}

type endpointerOptions struct {
	invalidateOnError bool
	invalidateTimeout time.Duration
}

// DefaultEndpointer implements an Endpointer interface.
// When created with NewEndpointer function, it automatically registers
// as a subscriber to events from the Instances and maintains a list
// of active Endpoints.
type DefaultEndpointer[Request, Response any] struct {
	cache     *endpointCache[Request, Response]
	instancer Instancer
	ch        chan Event
}

func (de *DefaultEndpointer[Request, Response]) receive() {
	for event := range de.ch {
		de.cache.Update(event)
	}
}

// Close deregisters DefaultEndpointer from the Instancer and stops the internal go-routine.
func (de *DefaultEndpointer[Request, Response]) Close() {
	de.instancer.Deregister(de.ch)
	close(de.ch)
}

// Endpoints implements Endpointer.
func (de *DefaultEndpointer[Request, Response]) Endpoints() ([]endpoint.Endpoint[Request, Response], error) {
	return de.cache.Endpoints()
}
