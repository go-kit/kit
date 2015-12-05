package zk

import (
	"github.com/eapache/channels"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/loadbalancer"
	"github.com/go-kit/kit/log"
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

	// try to create path nodes if they are not available
	p.client.CreateParentNodes(p.path)

	// intial node retrieval and cache fill
	instances, simpleOutChannel, err := p.client.GetEntries(p.path)
	if err != nil {
		logger.Log("path", p.path, "msg", "failed to retrieve entries", "err", err)
	} else {
		logger.Log("path", p.path, "instances", len(instances))
	}
	p.cache.Replace(instances)

	// handle incoming path updates
	go p.loop(simpleOutChannel)

	return p, nil
}

func (p *Publisher) loop(simpleOutChannel channels.SimpleOutChannel) {
	var (
		instances []string
		err       error
	)
	for {
		responseChan := simpleOutChannel.Out()
		select {
		case <-responseChan:
			// we received a path update notification, call GetEntries to
			// retrieve child node data and set new watch as zk watches are one
			// time triggers
			instances, simpleOutChannel, err = p.client.GetEntries(p.path)
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
