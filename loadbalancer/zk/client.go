package zk

import (
	"errors"
	"net"
	"strings"
	"time"

	"github.com/eapache/channels"
	"github.com/samuel/go-zookeeper/zk"

	"github.com/go-kit/kit/log"
)

// DefaultACL is the default ACL to use for creating znodes
var DefaultACL = zk.WorldACL(zk.PermAll)

const (
	// ConnectTimeout is the default timeout to establish a connection to a
	// ZooKeeper node.
	ConnectTimeout = 2 * time.Second
	// SessionTimeout is the default timeout to keep the current ZooKeeper
	// session alive during a temporary disconnect.
	SessionTimeout = 5 * time.Second
)

// Client is a wrapper around a lower level ZooKeeper client implementation.
type Client interface {
	// GetEntries should query the provided path in ZooKeeper, place a watch on
	// it and retrieve data from its current child nodes.
	GetEntries(path string) ([]string, channels.SimpleOutChannel, error)
	// CreateParentNodes should try to create the path in case it does not exist
	// yet on ZooKeeper.
	CreateParentNodes(path string) error
}

type clientConfig struct {
	acl             []zk.ACL
	connectTimeout  time.Duration
	sessionTimeout  time.Duration
	rootNodePayload [][]byte
}

type configOption func(c *clientConfig) error

type client struct {
	*zk.Conn
	clientConfig
	Eventc <-chan zk.Event
	quit   chan struct{}
}

// WithACL returns a configOption specifying a non-default ACL.
func WithACL(acl []zk.ACL) configOption {
	return func(c *clientConfig) error {
		c.acl = acl
		return nil
	}
}

// WithConnectTimeout returns a configOption specifying a non-default connection
// timeout when we try to establish a connection to a ZooKeeper server.
func WithConnectTimeout(t time.Duration) configOption {
	return func(c *clientConfig) error {
		if t.Seconds() < 1 {
			return errors.New("Invalid Connect Timeout. Minimum value is 1 second")
		}
		c.connectTimeout = t
		return nil
	}
}

// WithSessionTimeout returns a configOption specifying a non-default session
// timeout.
func WithSessionTimeout(t time.Duration) configOption {
	return func(c *clientConfig) error {
		if t.Seconds() < 1 {
			return errors.New("Invalid Session Timeout. Minimum value is 1 second")
		}
		c.sessionTimeout = t
		return nil
	}
}

// WithPayload returns a configOption specifying non-default data values for
// each znode created by CreateParentNodes.
func WithPayload(payload [][]byte) configOption {
	return func(c *clientConfig) error {
		c.rootNodePayload = payload
		return nil
	}
}

// NewClient returns a ZooKeeper client with a connection to the server cluster.
// It will return an error if the server cluster cannot be resolved.
func NewClient(servers []string, logger log.Logger, options ...configOption) (Client, error) {
	config := clientConfig{
		acl:            DefaultACL,
		connectTimeout: ConnectTimeout,
		sessionTimeout: SessionTimeout,
	}
	for _, option := range options {
		if err := option(&config); err != nil {
			return nil, err
		}
	}
	// dialer overrides the default ZooKeeper library Dialer so we can configure
	// the connectTimeout. The current library has a hardcoded value of 1 second
	// and there are reports of race conditions, due to slow DNS resolvers
	// and other network latency issues.
	dialer := func(network, address string, _ time.Duration) (net.Conn, error) {
		return net.DialTimeout(network, address, config.connectTimeout)
	}
	conn, eventc, err := zk.Connect(servers, config.sessionTimeout, withLogger(logger), zk.WithDialer(dialer))
	if err != nil {
		return nil, err
	}
	return &client{conn, config, eventc, make(chan struct{})}, nil
}

// CreateParentNodes implements the ZooKeeper Client interface.
func (c *client) CreateParentNodes(path string) error {
	payload := []byte("")
	pathString := ""
	pathNodes := strings.Split(path, "/")
	for i := 1; i < len(pathNodes); i++ {
		if i <= len(c.rootNodePayload) {
			payload = c.rootNodePayload[i-1]
		} else {
			payload = []byte("")
		}
		pathString += "/" + pathNodes[i]
		_, err := c.Create(pathString, payload, 0, c.acl)
		if err != nil && err != zk.ErrNodeExists && err != zk.ErrNoAuth {
			return err
		}
	}
	return nil
}

// GetEntries implements the ZooKeeper Client interface.
func (c *client) GetEntries(path string) ([]string, channels.SimpleOutChannel, error) {
	// retrieve list of child nodes for given path and add watch to path
	znodes, _, updateReceived, err := c.ChildrenW(path)
	if err != nil {
		return nil, channels.Wrap(updateReceived), err
	}

	var resp []string
	for _, znode := range znodes {
		// retrieve payload for child znode and add to response array
		if data, _, err := c.Get(path + "/" + znode); err == nil {
			resp = append(resp, string(data))
		}
	}
	return resp, channels.Wrap(updateReceived), nil
}
