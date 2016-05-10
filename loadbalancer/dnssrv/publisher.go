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
	name   string
	cache  *loadbalancer.EndpointCache
	logger log.Logger
	quit   chan struct{}
}

// NewPublisher returns a DNS SRV publisher. The name is resolved
// synchronously as part of construction; if that resolution fails, the
// constructor will return an error. The factory is used to convert a
// host:port to a usable endpoint. The logger is used to report DNS and
// factory errors.
func NewPublisher(
	name string,
	ttl time.Duration,
	factory loadbalancer.Factory,
	logger log.Logger,
) *Publisher {
	return NewPublisherDetailed(name, time.NewTicker(ttl), net.LookupSRV, factory, logger)
}

// NewPublisherDetailed is the same as NewPublisher, but allows users to provide
// an explicit lookup refresh ticker instead of a TTL, and specify the function
// used to perform lookups instead of using net.LookupSRV.
func NewPublisherDetailed(
	name string,
	refreshTicker *time.Ticker,
	lookupSRV func(service, proto, name string) (cname string, addrs []*net.SRV, err error),
	factory loadbalancer.Factory,
	logger log.Logger,
) *Publisher {
	p := &Publisher{
		name:   name,
		cache:  loadbalancer.NewEndpointCache(factory, logger),
		logger: logger,
		quit:   make(chan struct{}),
	}

	instances, err := p.resolve(lookupSRV)
	if err == nil {
		logger.Log("name", name, "instances", len(instances))
	} else {
		logger.Log("name", name, "err", err)
	}
	p.cache.Replace(instances)

	go p.loop(refreshTicker, lookupSRV)
	return p
}

// Stop terminates the publisher.
func (p *Publisher) Stop() {
	close(p.quit)
}

func (p *Publisher) loop(
	refreshTicker *time.Ticker,
	lookupSRV func(service, proto, name string) (cname string, addrs []*net.SRV, err error),
) {
	defer refreshTicker.Stop()
	for {
		select {
		case <-refreshTicker.C:
			instances, err := p.resolve(lookupSRV)
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
	return p.cache.Endpoints()
}

func (p *Publisher) resolve(lookupSRV func(service, proto, name string) (cname string, addrs []*net.SRV, err error)) ([]string, error) {
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
