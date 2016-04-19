package cache

import (
	"io"
	"sort"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/service"
)

// Cache collects the most recent set of services from a service discovery
// system via a subscriber.
type Cache struct {
	mtx     sync.RWMutex
	factory sd.Factory
	cache   map[string]serviceCloser
	logger  log.Logger
}

type serviceCloser struct {
	service.Service
	io.Closer
}

// New returns a new, empty service cache.
func New(factory sd.Factory, logger log.Logger) *Cache {
	return &Cache{
		factory: factory,
		cache:   map[string]serviceCloser{},
		logger:  logger,
	}
}

// Update should be invoked by clients with a complete set of current instance
// strings whenever that set changes. The cache manufactures new services via
// the factory, closes old services when they disappear, and persists existing
// services if they survive through an update.
func (c *Cache) Update(instances []string) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	// Produce the current set of services.
	cache := make(map[string]serviceCloser, len(instances))
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
		cache[instance] = serviceCloser{service, closer}
	}

	// Close any leftover services.
	for _, sc := range c.cache {
		if sc.Closer != nil {
			sc.Closer.Close()
		}
	}

	// Swap and trigger GC for old copy.
	c.cache = cache
}

// Services yields the current set of services, ordered lexicographically by the
// corresponding instance string.
func (c *Cache) Services() ([]service.Service, error) {
	c.mtx.RLock()
	defer c.mtx.RUnlock()

	slice := make([]instanceService, 0, len(c.cache))
	for instance, sc := range c.cache {
		slice = append(slice, instanceService{instance, sc.Service})
	}

	sort.Sort(byInstance(slice))

	services := make([]service.Service, len(slice))
	for i := range slice {
		services[i] = slice[i].service
	}

	return services, nil
}

func (c *Cache) len() int {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	return len(c.cache)
}

type instanceService struct {
	instance string
	service  service.Service
}

type byInstance []instanceService

func (a byInstance) Len() int           { return len(a) }
func (a byInstance) Less(i, j int) bool { return a[i].instance < a[j].instance }
func (a byInstance) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
