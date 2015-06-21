package loadbalancer

import (
	"crypto/md5"
	"fmt"
	"net"
	"sort"
	"time"

	"github.com/go-kit/kit/endpoint"
)

type dnssrvPublisher struct {
	subscribe   chan chan<- []endpoint.Endpoint
	unsubscribe chan chan<- []endpoint.Endpoint
	quit        chan struct{}
}

// NewDNSSRVPublisher returns a publisher that resolves the SRV name every ttl, and
func NewDNSSRVPublisher(name string, ttl time.Duration, makeEndpoint func(hostport string) endpoint.Endpoint) Publisher {
	p := &dnssrvPublisher{
		subscribe:   make(chan chan<- []endpoint.Endpoint),
		unsubscribe: make(chan chan<- []endpoint.Endpoint),
		quit:        make(chan struct{}),
	}
	go p.loop(name, ttl, makeEndpoint)
	return p
}

func (p *dnssrvPublisher) Subscribe(c chan<- []endpoint.Endpoint) {
	p.subscribe <- c
}

func (p *dnssrvPublisher) Unsubscribe(c chan<- []endpoint.Endpoint) {
	p.unsubscribe <- c
}

func (p *dnssrvPublisher) Stop() {
	close(p.quit)
}

var newTicker = time.NewTicker

func (p *dnssrvPublisher) loop(name string, ttl time.Duration, makeEndpoint func(hostport string) endpoint.Endpoint) {
	var (
		subscriptions = map[chan<- []endpoint.Endpoint]struct{}{}
		addrs, md5, _ = resolve(name)
		endpoints     = convert(addrs, makeEndpoint)
		ticker        = newTicker(ttl)
	)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			addrs, newmd5, err := resolve(name)
			if err == nil && newmd5 != md5 {
				endpoints = convert(addrs, makeEndpoint)
				for c := range subscriptions {
					c <- endpoints
				}
				md5 = newmd5
			}

		case c := <-p.subscribe:
			subscriptions[c] = struct{}{}
			c <- endpoints

		case c := <-p.unsubscribe:
			delete(subscriptions, c)

		case <-p.quit:
			return
		}
	}
}

// Allow mocking in tests.
var resolve = func(name string) (addrs []*net.SRV, md5sum string, err error) {
	_, addrs, err = net.LookupSRV("", "", name)
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

func convert(addrs []*net.SRV, makeEndpoint func(hostport string) endpoint.Endpoint) []endpoint.Endpoint {
	endpoints := make([]endpoint.Endpoint, len(addrs))
	for i, addr := range addrs {
		endpoints[i] = makeEndpoint(addr2hostport(addr))
	}
	return endpoints
}

func addr2hostport(addr *net.SRV) string {
	return net.JoinHostPort(addr.Target, fmt.Sprintf("%d", addr.Port))
}
