package strategy

import (
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/loadbalancer/publisher"
)

type cache struct {
	req  chan []endpoint.Endpoint
	cnt  chan int
	quit chan struct{}
}

func newCache(p publisher.Publisher) *cache {
	c := &cache{
		req:  make(chan []endpoint.Endpoint),
		cnt:  make(chan int),
		quit: make(chan struct{}),
	}
	go c.loop(p)
	return c
}

func (c *cache) loop(p publisher.Publisher) {
	e := make(chan []endpoint.Endpoint, 1)
	p.Subscribe(e)
	defer p.Unsubscribe(e)
	endpoints := <-e
	for {
		select {
		case endpoints = <-e:
		case c.cnt <- len(endpoints):
		case c.req <- endpoints:
		case <-c.quit:
			return
		}
	}
}

func (c *cache) count() int {
	return <-c.cnt
}

func (c *cache) get() []endpoint.Endpoint {
	return <-c.req
}

func (c *cache) stop() {
	close(c.quit)
}
