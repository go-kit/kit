package eureka

import (
	"fmt"

	fargo "github.com/hudl/fargo"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/cache"
)

// Subscriber yields endpoints stored in the Eureka registry for the given app.
// Changes in that app are watched and will update the Subscriber endpoints.
type Subscriber struct {
	client Client
	cache  *cache.Cache
	logger log.Logger
	app    string
	quitc  chan struct{}
}

var _ sd.Subscriber = &Subscriber{}

// NewSubscriber returns a Eureka subscriber. It will start watching the given
// app string for changes, and update the endpoints accordingly.
func NewSubscriber(c Client, factory sd.Factory, logger log.Logger, app string) *Subscriber {
	s := &Subscriber{
		client: c,
		cache:  cache.New(factory, logger),
		app:    app,
		logger: logger,
		quitc:  make(chan struct{}),
	}

	instances, err := s.getInstances()
	if err == nil {
		s.logger.Log("app", s.app, "instances", len(instances))
	} else {
		s.logger.Log("app", s.app, "msg", "failed to retrieve instances", "err", err)
	}

	s.cache.Update(instances)
	go s.loop()
	return s
}

func (s *Subscriber) getInstances() ([]string, error) {
	fargoInstances, err := s.client.Instances(s.app)
	if err != nil {
		return nil, err
	}
	return convertFargoInstances(fargoInstances), nil
}

func (s *Subscriber) loop() {
	updatec := s.client.ScheduleUpdates(s.app, s.quitc)
	for {
		select {
		case <-s.quitc:
			return
		case u := <-updatec:
			if u.Err != nil {
				s.logger.Log("app", s.app, "msg", "failed to retrieve instances", "err", u.Err)
				continue
			}

			instances := convertFargoApplication(u.App)
			s.logger.Log("app", s.app, "instances", len(instances))
			s.cache.Update(instances)
		}
	}
}

// Endpoints implements the Subscriber interface.
func (s *Subscriber) Endpoints() ([]endpoint.Endpoint, error) {
	return s.cache.Endpoints(), nil
}

// Stop terminates the Subscriber.
func (s *Subscriber) Stop() {
	close(s.quitc)
}

func convertFargoApplication(fargoApplication *fargo.Application) (instances []string) {
	if fargoApplication != nil {
		instances = convertFargoInstances(fargoApplication.Instances)
	}
	return instances
}

func convertFargoInstances(fargoInstances []*fargo.Instance) []string {
	instances := make([]string, len(fargoInstances))
	for i, fargoInstance := range fargoInstances {
		instances[i] = fmt.Sprintf("%s:%d", fargoInstance.IPAddr, fargoInstance.Port)
	}
	return instances
}
