package etcd

import (
	"time"

	"sync"

	"github.com/go-kit/kit/log"
)

// PeriodicRegistrar periodically registers service instance liveness information to etcd.
type PeriodicRegistrar struct {
	client     Client
	service    Service
	logger     log.Logger
	expiration time.Duration
	frequency  time.Duration
	stopC      chan chan bool
	mu         *sync.Mutex // mutex for registered
	registered bool
}

// NewPeriodicRegistrar returns a etcd Registrar with recurring registeation acting on the provided catalog
// registration (service).
func NewPeriodicRegistrar(client Client, service Service, logger log.Logger, regFrequency, regExpiration time.Duration) *PeriodicRegistrar {
	return &PeriodicRegistrar{
		client:  client,
		service: service,
		logger: log.NewContext(logger).With(
			"key", service.Key,
			"value", service.Value,
		),
		expiration: regExpiration,
		frequency:  regFrequency,
		stopC:      make(chan chan bool),
		mu:         &sync.Mutex{},
	}
}

// Register implements the sd.Registrar interface. Call it when you want your
// service to be registered in etcd, typically at startup.
func (r *PeriodicRegistrar) Register() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.registered {
		return
	}
	r.registered = true
	r.logger.Log("action", "register")
	r.register()
	go func() {
		tick := time.Tick(r.frequency)
		for {
			select {
			case <-tick:
				r.register()
			case doneC := <-r.stopC:
				r.deregister()
				doneC <- true
			}
		}
	}()
}

func (r *PeriodicRegistrar) register() {
	if err := r.client.RegisterTTL(r.service, r.expiration); err != nil {
		r.logger.Log("err", err)
	}
}

// Deregister implements the sd.Registrar interface. Call it when you want your
// service to be deregistered from etcd, typically just prior to shutdown.
func (r *PeriodicRegistrar) Deregister() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.registered {
		return
	}
	done := make(chan bool)
	r.stopC <- done
	<-done
	r.registered = false
}

func (r *PeriodicRegistrar) deregister() {
	if err := r.client.Deregister(r.service); err != nil {
		r.logger.Log("err", err)
	} else {
		r.logger.Log("action", "deregister")
	}
}
