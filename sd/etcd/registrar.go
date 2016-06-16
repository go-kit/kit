package etcd

import (
	etcd "github.com/coreos/etcd/client"
	"github.com/go-kit/kit/log"
)

// Registrar registers service instance liveness information to etcd.
type Registrar struct {
	client  Client
	service Service
	logger  log.Logger
}

// Service holds the key, value and instance identifying data you
// want to publish to etcd.
type Service struct {
	Key           string // discovery key, example: /myorganization/myplatform/
	Value         string // service name value, example: addsvc
	DeleteOptions *etcd.DeleteOptions
}

// NewRegistrar returns a etcd Registrar acting on the provided catalog
// registration.
func NewRegistrar(client Client, service Service, logger log.Logger) *Registrar {
	return &Registrar{
		client:  client,
		service: service,
		logger: log.NewContext(logger).With(
			"value", service.Value,
			"key", service.Key,
		),
	}
}

// Register implements sd.Registrar interface.
func (r *Registrar) Register() {
	if err := r.client.Register(r.service); err != nil {
		r.logger.Log("err", err)
	} else {
		r.logger.Log("action", "register")
	}
}

// Deregister implements sd.Registrar interface.
func (r *Registrar) Deregister() {
	if err := r.client.Deregister(r.service); err != nil {
		r.logger.Log("err", err)
	} else {
		r.logger.Log("action", "deregister")
	}
}
