package sd

import (
	"io"
	"sort"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/openmesh/kit/endpoint"
)

// endpointCache collects the most recent set of instances from a service discovery
// system, creates endpoints for them using a factory function, and makes
// them available to consumers.
type endpointCache[Request, Response any] struct {
	options            endpointerOptions
	mtx                sync.RWMutex
	factory            Factory[Request, Response]
	cache              map[string]endpointCloser[Request, Response]
	err                error
	endpoints          []endpoint.Endpoint[Request, Response]
	logger             log.Logger
	invalidateDeadline time.Time
	timeNow            func() time.Time
}

type endpointCloser[Request, Response any] struct {
	endpoint.Endpoint[Request, Response]
	io.Closer
}

// newEndpointCache returns a new, empty endpointCache.
func newEndpointCache[Request, Response any](factory Factory[Request, Response], logger log.Logger, options endpointerOptions) *endpointCache[Request, Response] {
	return &endpointCache[Request, Response]{
		options: options,
		factory: factory,
		cache:   map[string]endpointCloser[Request, Response]{},
		logger:  logger,
		timeNow: time.Now,
	}
}

// Update should be invoked by clients with a complete set of current instance
// strings whenever that set changes. The cache manufactures new endpoints via
// the factory, closes old endpoints when they disappear, and persists existing
// endpoints if they survive through an update.
func (c *endpointCache[Request, Response]) Update(event Event) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	// Happy path.
	if event.Err == nil {
		c.updateCache(event.Instances)
		c.err = nil
		return
	}

	// Sad path. Something's gone wrong in sd.
	c.logger.Log("err", event.Err)
	if !c.options.invalidateOnError {
		return // keep returning the last known endpoints on error
	}
	if c.err != nil {
		return // already in the error state, do nothing & keep original error
	}
	c.err = event.Err
	// set new deadline to invalidate Endpoints unless non-error Event is received
	c.invalidateDeadline = c.timeNow().Add(c.options.invalidateTimeout)
	return
}

func (c *endpointCache[Request, Response]) updateCache(instances []string) {
	// Deterministic order (for later).
	sort.Strings(instances)

	// Produce the current set of services.
	cache := make(map[string]endpointCloser[Request, Response], len(instances))
	for _, instance := range instances {
		// If it already exists, just copy it over.
		if sc, ok := c.cache[instance]; ok {
			cache[instance] = sc
			delete(c.cache, instance)
			continue
		}

		// If it doesn't exist, create it.
		service, closer, err := c.factory(instance)
		if err != nil {
			c.logger.Log("instance", instance, "err", err)
			continue
		}
		cache[instance] = endpointCloser[Request, Response]{service, closer}
	}

	// Close any leftover endpoints.
	for _, sc := range c.cache {
		if sc.Closer != nil {
			sc.Closer.Close()
		}
	}

	// Populate the slice of endpoints.
	endpoints := make([]endpoint.Endpoint[Request, Response], 0, len(cache))
	for _, instance := range instances {
		// A bad factory may mean an instance is not present.
		if _, ok := cache[instance]; !ok {
			continue
		}
		endpoints = append(endpoints, cache[instance].Endpoint)
	}

	// Swap and trigger GC for old copies.
	c.endpoints = endpoints
	c.cache = cache
}

// Endpoints yields the current set of (presumably identical) endpoints, ordered
// lexicographically by the corresponding instance string.
func (c *endpointCache[Request, Response]) Endpoints() ([]endpoint.Endpoint[Request, Response], error) {
	// in the steady state we're going to have many goroutines calling Endpoints()
	// concurrently, so to minimize contention we use a shared R-lock.
	c.mtx.RLock()

	if c.err == nil || c.timeNow().Before(c.invalidateDeadline) {
		defer c.mtx.RUnlock()
		return c.endpoints, nil
	}

	c.mtx.RUnlock()

	// in case of an error, switch to an exclusive lock.
	c.mtx.Lock()
	defer c.mtx.Unlock()

	// re-check condition due to a race between RUnlock() and Lock().
	if c.err == nil || c.timeNow().Before(c.invalidateDeadline) {
		return c.endpoints, nil
	}

	c.updateCache(nil) // close any remaining active endpoints
	return nil, c.err
}
