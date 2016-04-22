package consul

import consul "github.com/hashicorp/consul/api"

// Client is a wrapper around the Consul API.
type Client interface {
	Service(service string, tag string, queryOpts *consul.QueryOptions) ([]*consul.ServiceEntry, *consul.QueryMeta, error)
	Register(registration *consul.CatalogRegistration) error
	Deregister(registration *consul.CatalogRegistration) error
}

type client struct {
	consul *consul.Client
}

// NewClient returns an implementation of the Client interface expecting a fully
// setup Consul Client.
func NewClient(c *consul.Client) Client {
	return &client{
		consul: c,
	}
}

// GetInstances returns the list of healthy instances for a given service,
// filtered by tag.
func (c *client) Service(service, tag string, opts *consul.QueryOptions) ([]*consul.ServiceEntry, *consul.QueryMeta, error) {
	return c.consul.Health().Service(service, tag, true, opts)
}

func (c *client) Register(r *consul.CatalogRegistration) error {
	_, err := c.consul.Catalog().Register(r, &consul.WriteOptions{})
	return err
}

func (c *client) Deregister(r *consul.CatalogRegistration) error {
	_, err := c.consul.Catalog().Deregister(&consul.CatalogDeregistration{
		Node:       r.Node,
		Address:    r.Address,
		Datacenter: r.Datacenter,
		ServiceID:  r.Service.ID,
		CheckID:    r.Check.CheckID,
	}, &consul.WriteOptions{})
	return err
}
