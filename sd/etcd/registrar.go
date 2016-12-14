package etcd

import (
	etcd "github.com/coreos/etcd/client"

	"github.com/go-kit/kit/log"
	"time"
)

const (
	MinHeartBeatTime = time.Millisecond * 500
)

// Registrar registers service instance liveness information to etcd.
type Registrar struct {
	client  Client
	service Service
	logger  log.Logger
	quit    chan struct{}
}

// Service holds the instance identifying data you want to publish to etcd. Key
// must be unique, and value is the string returned to subscribers, typically
// called the "instance" string in other parts of package sd.
type Service struct {
	Key           string // unique key, e.g. "/service/foobar/1.2.3.4:8080"
	Value         string // returned to subscribers, e.g. "http://1.2.3.4:8080"
	TTL           *TTLOption
	DeleteOptions *etcd.DeleteOptions
}

// TTLOption allow setting a key with a TTL, and regularly refreshes the lease with a goroutine
type TTLOption struct {
	Heartbeat time.Duration
	TTL       time.Duration
}

// NewTTLOption returns a TTLOption
func NewTTLOption(heartbeat, ttl time.Duration) *TTLOption {
	if heartbeat <= MinHeartBeatTime {
		heartbeat = MinHeartBeatTime
	}
	if ttl <= heartbeat {
		ttl = heartbeat * 3
	}
	return &TTLOption{heartbeat, ttl}
}

// NewRegistrar returns a etcd Registrar acting on the provided catalog
// registration (service).
func NewRegistrar(client Client, service Service, logger log.Logger) *Registrar {
	return &Registrar{
		client:  client,
		service: service,
		logger: log.NewContext(logger).With(
			"key", service.Key,
			"value", service.Value,
		),
	}
}

// Register implements the sd.Registrar interface. Call it when you want your
// service to be registered in etcd, typically at startup.
func (r *Registrar) Register() {
	if err := r.client.Register(r.service); err != nil {
		r.logger.Log("err", err)
	} else {
		r.logger.Log("action", "register")
	}
	if r.service.TTL == nil {
		return
	}
	r.quit = make(chan struct{})
	go func() {
		for {
			select {
			case <-r.quit:
				return
			case <-time.After(r.service.TTL.Heartbeat):
				if err := r.client.Register(r.service); err != nil {
					r.logger.Log("err", err)
				}
			}
		}
	}()
}

// Deregister implements the sd.Registrar interface. Call it when you want your
// service to be deregistered from etcd, typically just prior to shutdown.
func (r *Registrar) Deregister() {
	if err := r.client.Deregister(r.service); err != nil {
		r.logger.Log("err", err)
	} else {
		r.logger.Log("action", "deregister")
	}
	if r.quit != nil {
		close(r.quit)
	}
}
