package consul

import (
	"fmt"
	"sync"
	"time"

	stdconsul "github.com/hashicorp/consul/api"

	"github.com/go-kit/kit/log"
)

// Registrar registers service instance liveness information to Consul.
type Registrar struct {
	client       Client
	registration *stdconsul.AgentServiceRegistration
	logger       log.Logger

	ttlCheckMtx  sync.Mutex
	ttlCheckQuit chan struct{}
}

// NewRegistrar returns a Consul Registrar acting on the provided catalog
// registration.
func NewRegistrar(client Client, r *stdconsul.AgentServiceRegistration, logger log.Logger) *Registrar {
	return &Registrar{
		client:       client,
		registration: r,
		logger:       log.With(logger, "service", r.Name, "tags", fmt.Sprint(r.Tags), "address", r.Address),
	}
}

// Register implements sd.Registrar interface.
func (p *Registrar) Register() {
	if err := p.client.Register(p.registration); err != nil {
		p.logger.Log("err", err)
	} else {
		p.logger.Log("action", "register")
	}
}

// Deregister implements sd.Registrar interface.
func (p *Registrar) Deregister() {
	if err := p.client.Deregister(p.registration); err != nil {
		p.logger.Log("err", err)
	} else {
		p.logger.Log("action", "deregister")
	}
}

// TTLCheck parameters for UpdateTTL.
type TTLCheck struct {
	output  string
	status  string        // "pass", "warn", "fail"
	timeout time.Duration // Lower than the AgentServiceCheck TTL
}

// AddCheck adds a service check to the Consul Registrar
// If a ttlCheck is provided, it spawns a goroutine to update the ttl.
func (p *Registrar) AddCheck(c *stdconsul.AgentCheckRegistration, ttl *TTLCheck) {
	if err := p.client.CheckRegister(c); err != nil {
		p.logger.Log("err", err)
	} else {
		p.logger.Log("action", "check added", "id", c.ID, "name", c.Name)
	}

	if ttl != nil {
		go p.ttlLoop(c.ID, ttl)
	}
}

// RemoveCheck removes a service check from the consul register
func (p *Registrar) RemoveCheck(checkID string) {
	if err := p.client.CheckDeregister(checkID); err != nil {
		p.logger.Log("err", err)
	} else {
		p.logger.Log("action", "check removed", "id", checkID)
	}

	p.ttlCheckMtx.Lock()
	if p.ttlCheckQuit != nil {
		close(p.ttlCheckQuit)
		p.ttlCheckQuit = nil
	}
	p.ttlCheckMtx.Unlock()
}

func (p *Registrar) ttlLoop(checkID string, ttl *TTLCheck) {
	p.ttlCheckMtx.Lock()
	if p.ttlCheckQuit != nil {
		return // already running
	}
	p.ttlCheckQuit = make(chan struct{})
	p.ttlCheckMtx.Unlock()

	tick := time.NewTicker(ttl.timeout)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			if err := p.client.UpdateTTL(checkID, ttl.output, ttl.status); err != nil {
				p.logger.Log("err", err)
			}
		case <-p.ttlCheckQuit:
			return
		}
	}
}
