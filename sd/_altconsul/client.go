package consul

import (
	"fmt"
	"net"

	consul "github.com/hashicorp/consul/api"
)

// Client is a wrapper around the Consul API.
type Client interface {
	// Register a service with the local agent.
	Register(r *consul.AgentServiceRegistration) error

	// Deregister a service with the local agent.
	Deregister(r *consul.AgentServiceRegistration) error

	// Prepare a catalog query. Must be deleted.
	Prepare(service string, tags []string, onlyHealthy bool) (queryID string, err error)

	// Read the set of instances from the prepared query.
	Read(queryID string) ([]string, error)

	// Delete a previously-prepared catalog query.
	Delete(queryID string) error
}

type client struct {
	consul *consul.Client
}

// NewClient returns an implementation of the Client interface, wrapping a
// concrete Consul client.
func NewClient(c *consul.Client) Client {
	return &client{
		consul: c,
	}
}

func (c *client) Register(r *consul.AgentServiceRegistration) error {
	return c.consul.Agent().ServiceRegister(r)
}

func (c *client) Deregister(r *consul.AgentServiceRegistration) error {
	return c.consul.Agent().ServiceDeregister(r.ID)
}

func (c *client) Prepare(service string, tags []string, onlyPassing bool) (string, error) {
	queryID, _, err := c.consul.PreparedQuery().Create(&consul.PreparedQueryDefinition{
		Service: consul.ServiceQuery{
			Service:     service, // only service is mandatory
			Tags:        tags,
			OnlyPassing: onlyPassing,
		},
	}, &consul.WriteOptions{})
	return queryID, err
}

func (c *client) Read(queryID string) ([]string, error) {
	resp, _, err := c.consul.PreparedQuery().Execute(queryID, &consul.QueryOptions{})
	if err != nil {
		return nil, err
	}

	instances := make([]string, len(resp.Nodes))
	for i, entry := range resp.Nodes {
		addr := entry.Node.Address
		if entry.Service.Address != "" {
			addr = entry.Service.Address
		}
		instances[i] = net.JoinHostPort(addr, fmt.Sprint(entry.Service.Port))
	}

	return instances, nil
}

func (c *client) Delete(queryID string) error {
	_, err := c.consul.PreparedQuery().Delete(queryID, &consul.QueryOptions{})
	return err
}
