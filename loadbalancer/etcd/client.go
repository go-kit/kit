package etcd

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	etcd "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
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
	Ctx context.Context
}

type ClientOptions struct {
	Cert                    string
	Key                     string
	CaCert                  string
	DialTimeout             time.Duration
	DialKeepAline           time.Duration
	HeaderTimeoutPerRequest time.Duration
}

// NewClient returns an *etcd.Client with a connection to the named machines.
// It will return an error if a connection to the cluster cannot be made.
// The parameter machines needs to be a full URL with schemas.
// e.g. "http://localhost:2379" will work, but "localhost:2379" will not.
func NewClient(ctx context.Context, machines []string, options *ClientOptions) (Client, error) {
	var (
		c        etcd.KeysAPI
		err      error
		caCertCt []byte
		tlsCert  tls.Certificate
	)
	if options == nil {
		options = &ClientOptions{
			Cert:                    "",
			Key:                     "",
			CaCert:                  "",
			DialTimeout:             100 * time.Second,
			DialKeepAline:           100 * time.Second,
			HeaderTimeoutPerRequest: 100 * time.Second,
		}
	}

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
					KeepAlive: options.DialKeepAline,
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
	resp, err := c.Get(c.Ctx, key, &etcd.GetOptions{Recursive: true})
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
		res, err := watch.Next(c.Ctx)
		if err != nil {
			return
		}
		responseChan <- res
	}
}
