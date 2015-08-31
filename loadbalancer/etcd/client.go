package etcd

import (
	"errors"
	"strings"

	"github.com/coreos/go-etcd/etcd"
)

// EtcdClient is a wrapper arround the etcd client
type EtcdClient interface {
	// GetEntries wil query the given prefix in etcd and returns a set of entries.
	GetEntries(prefix string) ([]string, error)
	// WatchPrefix starts watching every change for given prefix in etcd. When a
	// change is detected it will populate the responseChan when an *etcd.Response.
	WatchPrefix(prefix string, responseChan chan *etcd.Response)
}

type etcdClient struct {
	*etcd.Client
}

// NewClient returns an *etcd.Client with a connection to named machines.
// It will return an error if a connection to the cluster cannot be made.
func NewClient(machines []string, cert, key, caCert string) (*etcdClient, error) {
	var c *etcd.Client
	var err error

	if cert != "" && key != "" {
		c, err = etcd.NewTLSClient(machines, cert, key, caCert)
		if err != nil {
			return &etcdClient{c}, err
		}
	} else {
		c = etcd.NewClient(machines)
	}
	success := c.SetCluster(machines)
	if !success {
		return &etcdClient{c}, errors.New("cannot connect to the etcd cluster: " + strings.Join(machines, ","))
	}
	return &etcdClient{c}, nil
}

// GetEntries implements the EtcdClient interface.
func (c *etcdClient) GetEntries(key string) ([]string, error) {
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

// WatchPrefix implements the EtcdClient interface.
func (c *etcdClient) WatchPrefix(prefix string, watchChan chan *etcd.Response) {
	c.Watch(prefix, 0, true, watchChan, nil)
}

// FakeEtcdClient fakes an etcd client, used for testing.
type FakeEtcdClient struct {
	responses map[string]*etcd.Response
}

// GetEntries implements the EtcdClient interface.
func (c *FakeEtcdClient) GetEntries(prefix string) ([]string, error) {
	response, ok := c.responses[prefix]
	if !ok {
		return nil, errors.New("key not exist")
	}

	entries := make([]string, len(response.Node.Nodes))
	for i, node := range response.Node.Nodes {
		entries[i] = node.Value
	}
	return entries, nil
}

// WatchPrefix implements the EtcdClient interface.
func (c *FakeEtcdClient) WatchPrefix(prefix string, watchChan chan *etcd.Response) {}
