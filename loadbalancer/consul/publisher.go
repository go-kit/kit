package consul

import (
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/loadbalancer"
	"github.com/go-kit/kit/log"
)

// Publisher TODO
type Publisher struct {
	client      Client
	service     string
	tag         string
	passingOnly bool
	lastIndex   uint64
	cache       *loadbalancer.EndpointCache
	logger      log.Logger
	quit        chan struct{}
}

// NewPublisher TODO
func NewPublisher(c Client, service, tag string, passingOnly bool, f loadbalancer.Factory, logger log.Logger) *Publisher {
	p := &Publisher{
		client:      c,
		service:     service,
		tag:         tag,
		passingOnly: passingOnly,
		lastIndex:   0,
		cache:       loadbalancer.NewEndpointCache(f, logger),
		logger:      logger,
		quit:        make(chan struct{}),
	}

	instances, lastIndex, err := p.client.Service(service, tag, passingOnly, 0)
	if err == nil {
		logger.Log("service", p.service, "tag", p.tag, "instances", len(instances))
	} else {
		logger.Log("service", p.service, "tag", p.tag, "err", err)
	}
	p.lastIndex = lastIndex
	p.cache.Replace(instances)

	go p.loop()
	return p
}

// Endpoints TODO
func (p *Publisher) Endpoints() ([]endpoint.Endpoint, error) {
	return p.cache.Endpoints(), nil
}

// Stop TODO
func (p *Publisher) Stop() {
	close(p.quit)
}

func (p *Publisher) loop() {
	var (
		instances []string
		err       error
	)
	for {
		// TODO need a better API to interrupt the call in case of quit
		select {
		case <-p.quit:
			return
		default:
			instances, p.lastIndex, err = p.client.Service(p.service, p.tag, p.passingOnly, p.lastIndex)
			if err != nil {
				p.logger.Log("service", p.service, "tag", p.tag, "err", err)
				continue
			}
			p.cache.Replace(instances)
		}
	}
}
