package eureka

import (
	"fmt"

	"github.com/hudl/fargo"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/cache"
)

// Subscriber yields endpoints stored in the Eureka registry for the given app.
// Changes in that app are watched and will update the Subscriber endpoints.
type Subscriber struct {
	conn    fargoConnection
	app     string
	factory sd.Factory
	logger  log.Logger
	cache   *cache.Cache
	quitc   chan chan struct{}
}

var _ sd.Subscriber = (*Subscriber)(nil)

// NewSubscriber returns a Eureka subscriber. It will start watching the given
// app string for changes, and update the endpoints accordingly.
func NewSubscriber(conn fargoConnection, app string, factory sd.Factory, logger log.Logger) *Subscriber {
	logger = log.With(logger, "app", app)

	s := &Subscriber{
		conn:    conn,
		app:     app,
		factory: factory,
		logger:  logger,
		cache:   cache.New(factory, logger),
		quitc:   make(chan chan struct{}),
	}

	instances, err := s.getInstances()
	if err == nil {
		s.logger.Log("instances", len(instances))
	} else {
		s.logger.Log("during", "getInstances", "err", err)
	}

	s.cache.Update(instances)
	go s.loop()
	return s
}

// Endpoints implements the Subscriber interface.
func (s *Subscriber) Endpoints() ([]endpoint.Endpoint, error) {
	return s.cache.Endpoints(), nil
}

// Stop terminates the subscriber.
func (s *Subscriber) Stop() {
	q := make(chan struct{})
	s.quitc <- q
	<-q
	s.quitc = nil
}

func (s *Subscriber) loop() {
	var (
		await   = false
		done    = make(chan struct{})
		updatec = s.conn.ScheduleAppUpdates(s.app, await, done)
	)
	defer close(done)

	for {
		select {
		case update := <-updatec:
			if update.Err != nil {
				s.logger.Log("during", "Update", "err", update.Err)
				continue
			}
			instances := convertFargoAppToInstances(update.App)
			s.logger.Log("instances", len(instances))
			s.cache.Update(instances)

		case q := <-s.quitc:
			close(q)
			return
		}
	}
}

func (s *Subscriber) getInstances() ([]string, error) {
	app, err := s.conn.GetApp(s.app)
	if err != nil {
		return nil, err
	}
	return convertFargoAppToInstances(app), nil
}

func convertFargoAppToInstances(app *fargo.Application) []string {
	instances := make([]string, len(app.Instances))
	for i, inst := range app.Instances {
		instances[i] = fmt.Sprintf("%s:%d", inst.IPAddr, inst.Port)
	}
	return instances
}
