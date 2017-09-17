package instance

import (
	"reflect"
	"sort"
	"sync"

	"github.com/go-kit/kit/sd"
)

// Cache keeps track of resource instances provided to it via Update method
// and implements the Instancer interface
type Cache struct {
	mtx   sync.RWMutex
	state sd.Event
	reg   registry
}

// NewCache creates a new Cache.
func NewCache() *Cache {
	return &Cache{
		reg: registry{},
	}
}

// Update receives new instances from service discovery, stores them internally,
// and notifies all registered listeners.
func (c *Cache) Update(event sd.Event) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	sort.Strings(event.Instances)
	if reflect.DeepEqual(c.state, event) {
		return // no need to broadcast the same instances
	}

	c.state = event
	c.reg.broadcast(event)
}

// State returns the current state of discovery (instances or error) as sd.Event
func (c *Cache) State() sd.Event {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	return c.state
}

// Stop implements Instancer. Since the cache is just a plain-old store of data,
// Stop is a no-op.
func (c *Cache) Stop() {}

// Register implements Instancer.
func (c *Cache) Register(ch chan<- sd.Event) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.reg.register(ch)
	// always push the current state to new channels
	ch <- c.state
}

// Deregister implements Instancer.
func (c *Cache) Deregister(ch chan<- sd.Event) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.reg.deregister(ch)
}

// registry is not goroutine-safe.
type registry map[chan<- sd.Event]struct{}

func (r registry) broadcast(event sd.Event) {
	for c := range r {
		c <- event
	}
}

func (r registry) register(c chan<- sd.Event) {
	r[c] = struct{}{}
}

func (r registry) deregister(c chan<- sd.Event) {
	delete(r, c)
}
