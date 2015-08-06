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

func (p *Publisher) loop(endpoints []endpoint.Endpoint, md5 string) {
	t := newTicker(p.ttl)
	defer t.Stop()
	for {
		select {
		case p.endpoints <- endpoints:

		case <-t.C:
			// TODO should we do this out-of-band?
			addrs, newmd5, err := resolve(p.name)
			if err != nil {
				p.logger.Log("name", p.name, "err", err)
				continue // don't replace good endpoints with bad ones
			}
			if newmd5 == md5 {
				continue // no change
			}
			endpoints = makeEndpoints(addrs, p.factory, p.logger)
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
	hostports := make([]string, len(addrs))
	for i, addr := range addrs {
		hostports[i] = fmt.Sprintf("%s:%d", addr.Target, addr.Port)
	}
	sort.Sort(sort.StringSlice(hostports))
	h := md5.New()
	for _, hostport := range hostports {
		fmt.Fprintf(h, hostport)
	}
	return addrs, fmt.Sprintf("%x", h.Sum(nil)), nil
}

func makeEndpoints(addrs []*net.SRV, f loadbalancer.Factory, logger log.Logger) []endpoint.Endpoint {
	endpoints := make([]endpoint.Endpoint, 0, len(addrs))
	for _, addr := range addrs {
		endpoint, err := f(addr2instance(addr))
		if err != nil {
			logger.Log("instance", addr2instance(addr), "err", err)
			continue
		}
		endpoints = append(endpoints, endpoint)
	}
	return endpoints
}

func addr2instance(addr *net.SRV) string {
	return net.JoinHostPort(addr.Target, fmt.Sprint(addr.Port))
}
