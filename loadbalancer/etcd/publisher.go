package etcd

import (
	"github.com/coreos/go-etcd/etcd"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/loadbalancer"
	"github.com/go-kit/kit/log"
)

// Publisher yield endpoints stored in a certain etcd keyspace. Any kind of
// change in that keyspace is watched and wil update the Publisher endpoints.
type Publisher struct {
	client    Client
	prefix    string
	factory   loadbalancer.Factory
	logger    log.Logger
	endpoints chan []endpoint.Endpoint
	quit      chan struct{}
}

// NewPublisher returs a etcd publisher. Etcd will start watching the given
// prefix for changes and update the Publisher endpoints.
func NewPublisher(c Client, prefix string, f loadbalancer.Factory, logger log.Logger) (*Publisher, error) {
	logger = log.NewContext(logger).With("component", "Etcd Publisher")

	p := &Publisher{
		client:    c,
		prefix:    prefix,
		factory:   f,
		logger:    logger,
		endpoints: make(chan []endpoint.Endpoint),
		quit:      make(chan struct{}),
	}

	entries, err := p.client.GetEntries(prefix)
	if err != nil {
		return nil, err
	}
	go p.loop(makeEndpoints(entries, f, logger))
	return p, nil
}

func (p *Publisher) loop(endpoints []endpoint.Endpoint) {
	watchChan := make(chan *etcd.Response)
	go p.client.WatchPrefix(p.prefix, watchChan)

	for {
		select {
		case p.endpoints <- endpoints:

		case <-watchChan:
			entries, err := p.client.GetEntries(p.prefix)
			if err != nil {
				p.logger.Log("msg", "failed to retrieve entries", "err", err)
				continue
			}
			endpoints = makeEndpoints(entries, p.factory, p.logger)

		case <-p.quit:
			return
		}
	}
}

// Endpoints implements the Publisher interface.
func (p *Publisher) Endpoints() ([]endpoint.Endpoint, error) {
	select {
	case endpoints := <-p.endpoints:
		return endpoints, nil
	case <-p.quit:
		return nil, loadbalancer.ErrPublisherStopped
	}
}

// Stop terminates the publisher.
func (p *Publisher) Stop() {
	close(p.quit)
}

func makeEndpoints(addrs []string, f loadbalancer.Factory, logger log.Logger) []endpoint.Endpoint {
	endpoints := make([]endpoint.Endpoint, 0, len(addrs))

	for _, addr := range addrs {
		endpoint, err := f(addr)
		if err != nil {
			logger.Log("instance", addr, "err", err)
			continue
		}
		endpoints = append(endpoints, endpoint)
	}
	return endpoints
}
