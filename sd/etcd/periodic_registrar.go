package etcd

import (
	"time"

	"github.com/go-kit/kit/log"
)

// PeriodicRegistrar periodically registers service instance liveness information to etcd.
type PeriodicRegistrar struct {
	client     Client
	service    Service
	logger     log.Logger
	expiration time.Duration
	frequency  time.Duration
	stopC      chan bool
}

// NewPeriodicRegistrar returns a etcd Registrar with recurring registeation acting on the provided catalog
// registration (service).
func NewPeriodicRegistrar(client Client, service Service, logger log.Logger, frequencySec, expirationSec int) *PeriodicRegistrar {
	return &PeriodicRegistrar{
		client:  client,
		service: service,
		logger: log.NewContext(logger).With(
			"key", service.Key,
			"value", service.Value,
		),
		expiration: time.Duration(expirationSec) * time.Second,
		frequency:  time.Duration(frequencySec) * time.Second,
		stopC:      make(chan bool),
	}
}

// Register implements the sd.Registrar interface. Call it when you want your
// service to be registered in etcd, typically at startup.
func (r *PeriodicRegistrar) Register() {
	r.register()
	go func() {
		tick := time.Tick(r.frequency)
		for {
			select {
			case <-tick:
				r.register()
			case <-r.stopC:
				r.deregister()
			}
		}
	}()
}

func (r *PeriodicRegistrar) register() {
	r.stopC <- true
	if err := r.client.RegisterTTL(r.service, r.expiration); err != nil {
		r.logger.Log("err", err)
	} else {
		r.logger.Log("action", "register")
	}
}

// Deregister implements the sd.Registrar interface. Call it when you want your
// service to be deregistered from etcd, typically just prior to shutdown.
func (r *PeriodicRegistrar) Deregister() {
	r.stopC <- true
}

func (r *PeriodicRegistrar) deregister() {
	if err := r.client.Deregister(r.service); err != nil {
		r.logger.Log("err", err)
	} else {
		r.logger.Log("action", "deregister")
	}
}
