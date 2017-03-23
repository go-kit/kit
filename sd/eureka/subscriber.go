package eureka

import (
	"fmt"

	stdeureka "github.com/hudl/fargo"

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
func NewSubscriber(c Client, f sd.Factory, l log.Logger, app string) *Subscriber {
	s := &Subscriber{
		client: c,
		cache:  cache.New(f, l),
		app:    app,
		logger: l,
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
	stdInstances, err := s.client.Instances(s.app)
	if err != nil {
		return nil, err
	}
	return convertStdInstances(stdInstances), nil
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

			instances := convertStdApplication(u.App)
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

func convertStdApplication(stdApplication *stdeureka.Application) (instances []string) {
	if stdApplication != nil {
		instances = convertStdInstances(stdApplication.Instances)
	}
	return instances
}

func convertStdInstances(stdInstances []*stdeureka.Instance) []string {
	instances := make([]string, len(stdInstances))
	for i, stdInstance := range stdInstances {
		instances[i] = fmt.Sprintf("%s:%d", stdInstance.IPAddr, stdInstance.Port)
	}
	return instances
}
