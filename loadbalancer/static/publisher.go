package static

import (
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/loadbalancer"
	"github.com/go-kit/kit/loadbalancer/fixed"
	"github.com/go-kit/kit/log"
)

// Publisher yields a set of static endpoints as produced by the passed factory.
type Publisher struct{ publisher *fixed.Publisher }

// NewPublisher returns a static endpoint Publisher.
func NewPublisher(instances []string, factory loadbalancer.Factory, logger log.Logger) Publisher {
	logger = log.NewContext(logger).With("component", "Static Publisher")
	endpoints := []endpoint.Endpoint{}
	for _, instance := range instances {
		e, _, err := factory(instance) // never close
		if err != nil {
			logger.Log("instance", instance, "err", err)
			continue
		}
		endpoints = append(endpoints, e)
	}
	return Publisher{publisher: fixed.NewPublisher(endpoints)}
}

// Endpoints implements Publisher.
func (p Publisher) Endpoints() ([]endpoint.Endpoint, error) {
	return p.publisher.Endpoints()
}
