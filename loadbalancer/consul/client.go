package consul

import consul "github.com/hashicorp/consul/api"

// Client is a wrapper around the Consul API.
type Client interface {
	Service(service string, tag string, queryOpts *consul.QueryOptions) ([]*consul.ServiceEntry, *consul.QueryMeta, error)
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

// GetInstances returns the list of healthy entries for a given service filtered
// by tag.
func (c *client) Service(
	service string,
	tag string,
	opts *consul.QueryOptions,
) ([]*consul.ServiceEntry, *consul.QueryMeta, error) {
	return c.consul.Health().Service(service, tag, true, opts)
}
