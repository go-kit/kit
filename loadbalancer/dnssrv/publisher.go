package dnssrv

import (
	"crypto/md5"
	"fmt"
	"net"
	"sort"
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
	factory   loadbalancer.Factory
	logger    log.Logger
	endpoints chan []endpoint.Endpoint
	quit      chan struct{}
}

// NewPublisher returns a DNS SRV publisher. The name is resolved
// synchronously as part of construction; if that resolution fails, the
// constructor will return an error. The factory is used to convert a
// host:port to a usable endpoint. The logger is used to report DNS and
// factory errors.
func NewPublisher(name string, ttl time.Duration, f loadbalancer.Factory, logger log.Logger) (*Publisher, error) {
	logger = log.NewContext(logger).With("component", "DNS SRV Publisher")
	addrs, md5, err := resolve(name)
	if err != nil {
		return nil, err
	}
	p := &Publisher{
		name:      name,
		ttl:       ttl,
		factory:   f,
		logger:    logger,
		endpoints: make(chan []endpoint.Endpoint),
		quit:      make(chan struct{}),
	}
	go p.loop(makeEndpoints(addrs, f, logger), md5)
	return p, nil
}

// Stop terminates the publisher.
func (p *Publisher) Stop() {
	close(p.quit)
}

func (p *Publisher) loop(m map[string]endpointCloser, md5 string) {
	t := newTicker(p.ttl)
	defer t.Stop()
	for {
		select {
		case p.endpoints <- flatten(m):

		case <-t.C:
			// TODO should we do this out-of-band?
			addrs, newmd5, err := resolve(p.name)
			if err != nil {
				p.logger.Log("name", p.name, "err", err)
				continue // don't replace probably-good endpoints with bad ones
			}
			if newmd5 == md5 {
				continue // optimization: no change
			}
			m = migrate(m, makeEndpoints(addrs, p.factory, p.logger))
			md5 = newmd5

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

func resolve(name string) (addrs []*net.SRV, md5sum string, err error) {
	_, addrs, err = lookupSRV("", "", name)
	if err != nil {
		return addrs, "", err
	}
	instances := make([]string, len(addrs))
	for i, addr := range addrs {
		instances[i] = addr2instance(addr)
	}
	sort.Sort(sort.StringSlice(instances))
	h := md5.New()
	for _, instance := range instances {
		fmt.Fprintf(h, instance)
	}
	return addrs, fmt.Sprintf("%x", h.Sum(nil)), nil
}

func makeEndpoints(addrs []*net.SRV, f loadbalancer.Factory, logger log.Logger) map[string]endpointCloser {
	m := make(map[string]endpointCloser, len(addrs))
	for _, addr := range addrs {
		instance := addr2instance(addr)
		endpoint, closer, err := f(instance)
		if err != nil {
			logger.Log("instance", addr2instance(addr), "err", err)
			continue
		}
		m[instance] = endpointCloser{endpoint, closer}
	}
	return m
}

func migrate(prev, curr map[string]endpointCloser) map[string]endpointCloser {
	for instance, ec := range prev {
		if _, ok := curr[instance]; !ok {
			close(ec.Closer)
		}
	}
	return curr
}

func addr2instance(addr *net.SRV) string {
	return net.JoinHostPort(addr.Target, fmt.Sprint(addr.Port))
}

func flatten(m map[string]endpointCloser) []endpoint.Endpoint {
	a := make([]endpoint.Endpoint, 0, len(m))
	for _, ec := range m {
		a = append(a, ec.Endpoint)
	}
	return a
}

type endpointCloser struct {
	endpoint.Endpoint
	loadbalancer.Closer
}
