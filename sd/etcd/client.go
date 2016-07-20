package etcd

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	etcd "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

var (
	ErrNoKey   = errors.New("no key provided")
	ErrNoValue = errors.New("no value provided")
)

// Client is a wrapper around the etcd client.
type Client interface {
	// GetEntries will query the given prefix in etcd and returns a set of entries.
	GetEntries(prefix string) ([]string, error)

	// WatchPrefix starts watching every change for given prefix in etcd. When an
	// change is detected it will populate the responseChan when an *etcd.Response.
	WatchPrefix(prefix string, responseChan chan *etcd.Response)

	// Register a service with etcd.
	Register(s Service) error
	// Deregister a service with etcd.
	Deregister(s Service) error
}

type client struct {
	keysAPI etcd.KeysAPI
	ctx     context.Context
}

// ClientOptions defines options for the etcd client.
type ClientOptions struct {
	Cert                    string
	Key                     string
	CaCert                  string
	DialTimeout             time.Duration
	DialKeepAlive           time.Duration
	HeaderTimeoutPerRequest time.Duration
}

// NewClient returns an *etcd.Client with a connection to the named machines.
// It will return an error if a connection to the cluster cannot be made.
// The parameter machines needs to be a full URL with schemas.
// e.g. "http://localhost:2379" will work, but "localhost:2379" will not.
func NewClient(ctx context.Context, machines []string, options ClientOptions) (Client, error) {
	var (
		c        etcd.KeysAPI
		err      error
		caCertCt []byte
		tlsCert  tls.Certificate
	)

	if options.Cert != "" && options.Key != "" {
		tlsCert, err = tls.LoadX509KeyPair(options.Cert, options.Key)
		if err != nil {
			return nil, err
		}

		caCertCt, err = ioutil.ReadFile(options.CaCert)
		if err != nil {
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCertCt)

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{tlsCert},
			RootCAs:      caCertPool,
		}

		transport := &http.Transport{
			TLSClientConfig: tlsConfig,
			Dial: func(network, addr string) (net.Conn, error) {
				dial := &net.Dialer{
					Timeout:   options.DialTimeout,
					KeepAlive: options.DialKeepAlive,
				}
				return dial.Dial(network, addr)
			},
		}

		cfg := etcd.Config{
			Endpoints:               machines,
			Transport:               transport,
			HeaderTimeoutPerRequest: options.HeaderTimeoutPerRequest,
		}
		ce, err := etcd.New(cfg)
		if err != nil {
			return nil, err
		}
		c = etcd.NewKeysAPI(ce)
	} else {
		cfg := etcd.Config{
			Endpoints:               machines,
			Transport:               etcd.DefaultTransport,
			HeaderTimeoutPerRequest: options.HeaderTimeoutPerRequest,
		}
		ce, err := etcd.New(cfg)
		if err != nil {
			return nil, err
		}
		c = etcd.NewKeysAPI(ce)
	}

	return &client{c, ctx}, nil
}

// GetEntries implements the etcd Client interface.
func (c *client) GetEntries(key string) ([]string, error) {
	resp, err := c.keysAPI.Get(c.ctx, key, &etcd.GetOptions{Recursive: true})
	if err != nil {
		return nil, err
	}

	// Special case. Note that it's possible that len(resp.Node.Nodes) == 0 and
	// resp.Node.Value is also empty, in which case the key is empty and we
	// should not return any entries.
	if len(resp.Node.Nodes) == 0 && resp.Node.Value != "" {
		return []string{resp.Node.Value}, nil
	}

	entries := make([]string, len(resp.Node.Nodes))
	for i, node := range resp.Node.Nodes {
		entries[i] = node.Value
	}
	return entries, nil
}

// WatchPrefix implements the etcd Client interface.
func (c *client) WatchPrefix(prefix string, responseChan chan *etcd.Response) {
	watch := c.keysAPI.Watcher(prefix, &etcd.WatcherOptions{AfterIndex: 0, Recursive: true})
	responseChan <- nil // TODO(pb) explain this
	for {
		res, err := watch.Next(c.ctx)
		if err != nil {
			return
		}
		responseChan <- res
	}
}

func (c *client) Register(s Service) error {
	if s.Key == "" {
		return ErrNoKey
	}
	if s.Value == "" {
		return ErrNoValue
	}
	_, err := c.keysAPI.Create(c.ctx, s.Key, s.Value)
	return err
}

func (c *client) Deregister(s Service) error {
	if s.Key == "" {
		return ErrNoKey
	}
	_, err := c.keysAPI.Delete(c.ctx, s.Key, s.DeleteOptions)
	return err
}
