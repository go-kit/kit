package etcd

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
	etcd "github.com/coreos/etcd/client"
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
	etcd.KeysAPI
}

// NewClient returns an *etcd.Client with a connection to the named machines.
// It will return an error if a connection to the cluster cannot be made.
// The parameter machines needs to be a full URL with schemas.
// e.g. "http://localhost:4001" will work, but "localhost:4001" will not.
func NewClient(machines []string, cert, key, caCert string) (Client, error) {
	var c etcd.KeysAPI

	if cert != "" && key != "" {

		tlsCert, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			return nil, err
		}

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{tlsCert},
			//			InsecureSkipVerify: true,
		}

		transport := &http.Transport{
			TLSClientConfig: tlsConfig,
			Dial: func(network, addr string) (net.Conn, error) {
				dialer := net.Dialer{
					Timeout:   time.Second,
					KeepAlive: time.Second,
				}

				return dialer.Dial(network, addr)
			},
		}

		cfg := etcd.Config{
			Endpoints:               machines,
			Transport:               transport,
			HeaderTimeoutPerRequest: time.Second,
		}
		ce, err := etcd.New(cfg)
		if err != nil {
			return nil, err
		}
		c = etcd.NewKeysAPI(ce)

	} else {
		//		c = etcd.NewClient(machines)
		cfg := etcd.Config{
			Endpoints:               machines,
			Transport:               etcd.DefaultTransport,
			HeaderTimeoutPerRequest: time.Second,
		}
		ce, err := etcd.New(cfg)
		if err != nil {
			return nil, err
		}
		c = etcd.NewKeysAPI(ce)
	}
	return &client{c}, nil
}

// GetEntries implements the etcd Client interface.
func (c *client) GetEntries(key string) ([]string, error) {
	resp, err := c.Get(context.Background(), key, &etcd.GetOptions{Recursive: true})
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
	watch := c.Watcher(prefix, &etcd.WatcherOptions{AfterIndex: 0, Recursive: true})
	for {
		res, err := watch.Next(context.Background())
		if err != nil {
			return
		}
		responseChan <- res
	}
}
