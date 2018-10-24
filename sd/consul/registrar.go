package consul

import (
	"fmt"

	"github.com/go-kit/kit/log"
	stdconsul "github.com/hashicorp/consul/api"
)

// Registrar registers service instance liveness information to Consul.
type Registrar struct {
	client       Client
	registration *stdconsul.AgentServiceRegistration
	logger       log.Logger
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
func (p *Registrar) Register() error {
	err := p.client.Register(p.registration)
	if err != nil {
		p.logger.Log("err", err)
		return err
	}
	p.logger.Log("action", "register")
	return nil
}

// Deregister implements sd.Registrar interface.
func (p *Registrar) Deregister() error {
	err := p.client.Deregister(p.registration)
	if err != nil {
		p.logger.Log("err", err)
		return err
	}
	p.logger.Log("action", "deregister")
	return nil
}
