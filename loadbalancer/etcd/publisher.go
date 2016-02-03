package etcd

import (
	etcd "github.com/coreos/etcd/client"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/loadbalancer"
	"github.com/go-kit/kit/log"
)

// Publisher yield endpoints stored in a certain etcd keyspace. Any kind of
// change in that keyspace is watched and will update the Publisher endpoints.
type Publisher struct {
	client Client
	prefix string
	cache  *loadbalancer.EndpointCache
	logger log.Logger
	quit   chan struct{}
}

// NewPublisher returs a etcd publisher. Etcd will start watching the given
// prefix for changes and update the Publisher endpoints.
func NewPublisher(c Client, prefix string, f loadbalancer.Factory, logger log.Logger) (*Publisher, error) {
	p := &Publisher{
		client: c,
		prefix: prefix,
		cache:  loadbalancer.NewEndpointCache(f, logger),
		logger: logger,
		quit:   make(chan struct{}),
	}

	instances, err := p.client.GetEntries(p.prefix)
	if err == nil {
		logger.Log("prefix", p.prefix, "instances", len(instances))
	} else {
		logger.Log("prefix", p.prefix, "err", err)
	}
	p.cache.Replace(instances)

	go p.loop()
	return p, nil
}

func (p *Publisher) loop() {
	responseChan := make(chan *etcd.Response)
	go p.client.WatchPrefix(p.prefix, responseChan)
	for {
		select {
		case <-responseChan:
			instances, err := p.client.GetEntries(p.prefix)
			if err != nil {
				p.logger.Log("msg", "failed to retrieve entries", "err", err)
				continue
			}
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
