package consul

import (
	"fmt"
	"strings"

	consul "github.com/hashicorp/consul/api"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/internal/cache"
	"github.com/go-kit/kit/service"
)

const defaultIndex = 0

// Subscriber yields endpoints for a service in Consul. Updates to the service
// are watched and will update the Subscriber endpoints.
type Subscriber struct {
	cache      *cache.Cache
	client     Client
	logger     log.Logger
	service    string
	tags       []string
	endpointsc chan []endpoint.Endpoint
	quitc      chan struct{}
}

var _ sd.Subscriber = &Subscriber{}

// NewSubscriber returns a Consul subscriber which returns Endpoints for the
// requested service. It only returns instances for which all of the passed tags
// are present.
func NewSubscriber(client Client, factory sd.Factory, logger log.Logger, service string, tags ...string) (*Subscriber, error) {
	s := &Subscriber{
		cache:   cache.New(factory, logger),
		client:  client,
		logger:  logger,
		service: service,
		tags:    tags,
		quitc:   make(chan struct{}),
	}

	instances, index, err := s.getInstances(defaultIndex)
	if err == nil {
		logger.Log("service", service, "tags", strings.Join(tags, ", "), "instances", len(instances))
	} else {
		logger.Log("service", service, "tags", strings.Join(tags, ", "), "err", err)
	}
	s.cache.Update(instances)

	go s.loop(index)

	return s, nil
}

// Services implements the Subscriber interface.
func (s *Subscriber) Services() ([]service.Service, error) {
	return s.cache.Services(), nil
}

// Stop terminates the subscriber.
func (s *Subscriber) Stop() {
	close(s.quitc)
}

func (s *Subscriber) loop(lastIndex uint64) {
	var (
		errc = make(chan error, 1)
		resc = make(chan response, 1)
	)

	for {
		go func() {
			instances, index, err := s.getInstances(lastIndex)
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
			s.logger.Log("service", s.service, "err", err)
		case res := <-resc:
			s.cache.Update(res.instances)
			lastIndex = res.index
		case <-s.quitc:
			return
		}
	}
}

func (s *Subscriber) getInstances(lastIndex uint64) ([]string, uint64, error) {
	tag := ""

	if len(s.tags) > 0 {
		tag = s.tags[0]
	}

	entries, meta, err := s.client.Service(s.service, tag, &consul.QueryOptions{
		WaitIndex: lastIndex,
	})
	if err != nil {
		return nil, 0, err
	}

	// If more than one tag is passed we need to filter it in the subscriber
	// until Consul supports multiple tags.
	// https://github.com/hashicorp/consul/issues/294
	if len(s.tags) > 1 {
		entries = filterEntries(entries, s.tags[1:]...)
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
