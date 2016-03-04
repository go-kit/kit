package zk

import (
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/loadbalancer"
	"github.com/go-kit/kit/log"
	"github.com/samuel/go-zookeeper/zk"
)

// Publisher yield endpoints stored in a certain ZooKeeper path. Any kind of
// change in that path is watched and will update the Publisher endpoints.
type Publisher struct {
	client Client
	path   string
	cache  *loadbalancer.EndpointCache
	logger log.Logger
	quit   chan struct{}
}

// NewPublisher returns a ZooKeeper publisher. ZooKeeper will start watching the
// given path for changes and update the Publisher endpoints.
func NewPublisher(c Client, path string, f loadbalancer.Factory, logger log.Logger) (*Publisher, error) {
	p := &Publisher{
		client: c,
		path:   path,
		cache:  loadbalancer.NewEndpointCache(f, logger),
		logger: logger,
		quit:   make(chan struct{}),
	}

	err := p.client.CreateParentNodes(p.path)
	if err != nil {
		return nil, err
	}

	// initial node retrieval and cache fill
	instances, eventc, err := p.client.GetEntries(p.path)
	if err != nil {
		logger.Log("path", p.path, "msg", "failed to retrieve entries", "err", err)
		return nil, err
	}
	logger.Log("path", p.path, "instances", len(instances))
	p.cache.Replace(instances)

	// handle incoming path updates
	go p.loop(eventc)

	return p, nil
}

func (p *Publisher) loop(eventc <-chan zk.Event) {
	var (
		instances []string
		err       error
	)
	for {
		select {
		case <-eventc:
			// we received a path update notification, call GetEntries to
			// retrieve child node data and set new watch as zk watches are one
			// time triggers
			instances, eventc, err = p.client.GetEntries(p.path)
			if err != nil {
				p.logger.Log("path", p.path, "msg", "failed to retrieve entries", "err", err)
				continue
			}
			p.logger.Log("path", p.path, "instances", len(instances))
			p.cache.Replace(instances)
		case <-p.quit:
			return
		}
	}
}

// Endpoints implements the Publisher interface.
func (p *Publisher) Endpoints() ([]endpoint.Endpoint, error) {
	return p.cache.Endpoints()
}

// Stop terminates the Publisher.
func (p *Publisher) Stop() {
	close(p.quit)
}
