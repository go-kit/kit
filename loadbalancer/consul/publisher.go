package consul

import (
	"fmt"
	"strings"

	consul "github.com/hashicorp/consul/api"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/loadbalancer"
	"github.com/go-kit/kit/log"
)

const defaultIndex = 0

// Publisher yields endpoints for a service in Consul. Updates to the service
// are watched and will update the Publisher endpoints.
type Publisher struct {
	cache      *loadbalancer.EndpointCache
	client     Client
	logger     log.Logger
	service    string
	tags       []string
	endpointsc chan []endpoint.Endpoint
	quitc      chan struct{}
}

// NewPublisher returns a Consul publisher which returns Endpoints for the
// requested service. It only returns instances for which all of the passed
// tags are present.
func NewPublisher(
	client Client,
	factory loadbalancer.Factory,
	logger log.Logger,
	service string,
	tags ...string,
) (*Publisher, error) {
	p := &Publisher{
		cache:   loadbalancer.NewEndpointCache(factory, logger),
		client:  client,
		logger:  logger,
		service: service,
		tags:    tags,
		quitc:   make(chan struct{}),
	}

	instances, index, err := p.getInstances(defaultIndex)
	if err == nil {
		logger.Log("service", service, "tags", strings.Join(tags, ", "), "instances", len(instances))
	} else {
		logger.Log("service", service, "tags", strings.Join(tags, ", "), "err", err)
	}
	p.cache.Replace(instances)

	go p.loop(index)

	return p, nil
}

// Endpoints implements the Publisher interface.
func (p *Publisher) Endpoints() ([]endpoint.Endpoint, error) {
	return p.cache.Endpoints()
}

// Stop terminates the publisher.
func (p *Publisher) Stop() {
	close(p.quitc)
}

func (p *Publisher) loop(lastIndex uint64) {
	var (
		errc = make(chan error, 1)
		resc = make(chan response, 1)
	)

	for {
		go func() {
			instances, index, err := p.getInstances(lastIndex)
			if err != nil {
				errc <- err
				return
			}
			resc <- response{
				index:     index,
				instances: instances,
			}
		}()

		select {
		case err := <-errc:
			p.logger.Log("service", p.service, "err", err)
		case res := <-resc:
			p.cache.Replace(res.instances)
			lastIndex = res.index
		case <-p.quitc:
			return
		}
	}
}

func (p *Publisher) getInstances(lastIndex uint64) ([]string, uint64, error) {
	tag := ""

	if len(p.tags) > 0 {
		tag = p.tags[0]
	}

	entries, meta, err := p.client.Service(
		p.service,
		tag,
		&consul.QueryOptions{
			WaitIndex: lastIndex,
		},
	)
	if err != nil {
		return nil, 0, err
	}

	// If more than one tag is passed we need to filter it in the publisher until
	// Consul supports multiple tags[0].
	//
	// [0] https://github.com/hashicorp/consul/issues/294
	if len(p.tags) > 1 {
		entries = filterEntries(entries, p.tags[1:]...)
	}

	return makeInstances(entries), meta.LastIndex, nil
}

// response is used as container to transport instances as well as the updated
// index.
type response struct {
	index     uint64
	instances []string
}

func filterEntries(entries []*consul.ServiceEntry, tags ...string) []*consul.ServiceEntry {
	var es []*consul.ServiceEntry

ENTRIES:
	for _, entry := range entries {
		ts := make(map[string]struct{}, len(entry.Service.Tags))

		for _, tag := range entry.Service.Tags {
			ts[tag] = struct{}{}
		}

		for _, tag := range tags {
			if _, ok := ts[tag]; !ok {
				continue ENTRIES
			}
		}

		es = append(es, entry)
	}

	return es
}

func makeInstances(entries []*consul.ServiceEntry) []string {
	instances := make([]string, len(entries))

	for i, entry := range entries {
		addr := entry.Node.Address

		if entry.Service.Address != "" {
			addr = entry.Service.Address
		}

		instances[i] = fmt.Sprintf("%s:%d", addr, entry.Service.Port)
	}

	return instances
}
