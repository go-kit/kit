package dnssrv

import (
	"fmt"
	"net"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/loadbalancer"
	"github.com/go-kit/kit/log"
)

// Publisher yields endpoints taken from the named DNS SRV record. The name is
// resolved on a fixed schedule. Priorities and weights are ignored.
type Publisher struct {
	name      string
	ttl       time.Duration
	cache     *loadbalancer.EndpointCache
	logger    log.Logger
	endpoints chan []endpoint.Endpoint
	quit      chan struct{}
}

// NewPublisher returns a DNS SRV publisher. The name is resolved
// synchronously as part of construction; if that resolution fails, the
// constructor will return an error. The factory is used to convert a
// host:port to a usable endpoint. The logger is used to report DNS and
// factory errors.
func NewPublisher(name string, ttl time.Duration, factory loadbalancer.Factory, logger log.Logger) *Publisher {
	p := &Publisher{
		name:      name,
		ttl:       ttl,
		cache:     loadbalancer.NewEndpointCache(factory, logger),
		logger:    logger,
		endpoints: make(chan []endpoint.Endpoint),
		quit:      make(chan struct{}),
	}

	instances, err := p.resolve()
	if err != nil {
		logger.Log(name, len(instances))
	} else {
		logger.Log(name, err)
	}
	p.cache.Replace(instances)

	go p.loop()
	return p
}

// Stop terminates the publisher.
func (p *Publisher) Stop() {
	close(p.quit)
}

func (p *Publisher) loop() {
	t := newTicker(p.ttl)
	defer t.Stop()
	for {
		select {
		case p.endpoints <- p.cache.Endpoints():

		case <-t.C:
			instances, err := p.resolve()
			if err != nil {
				p.logger.Log(p.name, err)
				continue // don't replace potentially-good with bad
			}
			p.cache.Replace(instances)

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

var (
	lookupSRV = net.LookupSRV
	newTicker = time.NewTicker
)

func (p *Publisher) resolve() ([]string, error) {
	_, addrs, err := lookupSRV("", "", p.name)
	if err != nil {
		return []string{}, err
	}
	instances := make([]string, len(addrs))
	for i, addr := range addrs {
		instances[i] = net.JoinHostPort(addr.Target, fmt.Sprint(addr.Port))
	}
	return instances, nil
}
