package etcd

import (
	"fmt"
	"strings"

	"github.com/coreos/go-etcd/etcd"
)

// Client is a wrapper arround the etcd client.
type Client interface {
	// GetEntries will query the given prefix in etcd and returns a set of entries.
	GetEntries(prefix string) ([]string, error)
	// WatchPrefix starts watching every change for given prefix in etcd. When an
	// change is detected it will populate the responseChan when an *etcd.Response.
	WatchPrefix(prefix string, responseChan chan *etcd.Response)
}

type client struct {
	*etcd.Client
}

// NewClient returns an *etcd.Client with a connection to the named machines.
// It will return an error if a connection to the cluster cannot be made.
// The parameter machines needs to be a full URL with schemas.
// e.g. "http://localhost:4001" will work, but "localhost:4001" will not.
func NewClient(machines []string, cert, key, caCert string) (Client, error) {
	var c *etcd.Client
	var err error

	if cert != "" && key != "" {
		c, err = etcd.NewTLSClient(machines, cert, key, caCert)
		if err != nil {
			return nil, err
		}
	} else {
		c = etcd.NewClient(machines)
	}
	success := c.SetCluster(machines)
	if !success {
		return nil, fmt.Errorf("cannot connect to the etcd cluster: %s", strings.Join(machines, ","))
	}
	return &client{c}, nil
}

// GetEntries implements the etcd Client interface.
func (c *client) GetEntries(key string) ([]string, error) {
	resp, err := c.Get(key, false, true)
	if err != nil {
		return nil, err
	}

	entries := make([]string, len(resp.Node.Nodes))
	for i, node := range resp.Node.Nodes {
		entries[i] = node.Value
	}
	return entries, nil
}

// WatchPrefix implements the etcd Client interface.
func (c *client) WatchPrefix(prefix string, responseChan chan *etcd.Response) {
	c.Watch(prefix, 0, true, responseChan, nil)
}
