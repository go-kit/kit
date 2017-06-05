package eureka

import (
	"fmt"

	"github.com/hudl/fargo"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/internal/instance"
)

// Instancer yields instances stored in the Eureka registry for the given app.
// Changes in that app are watched and will update the subscribers.
type Instancer struct {
	cache  *instance.Cache
	conn   fargoConnection
	app    string
	logger log.Logger
	quitc  chan chan struct{}
}

// NewInstancer returns a Eureka Instancer. It will start watching the given
// app string for changes, and update the subscribers accordingly.
func NewInstancer(conn fargoConnection, app string, logger log.Logger) *Instancer {
	logger = log.With(logger, "app", app)

	s := &Instancer{
		cache:  instance.NewCache(),
		conn:   conn,
		app:    app,
		logger: logger,
		quitc:  make(chan chan struct{}),
	}

	instances, err := s.getInstances()
	if err == nil {
		s.logger.Log("instances", len(instances))
	} else {
		s.logger.Log("during", "getInstances", "err", err)
	}

	s.cache.Update(sd.Event{Instances: instances, Err: err})
	go s.loop()
	return s
}

// Stop terminates the Instancer.
func (s *Instancer) Stop() {
	q := make(chan struct{})
	s.quitc <- q
	<-q
	s.quitc = nil
}

func (s *Instancer) loop() {
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
				s.cache.Update(sd.Event{Err: update.Err})
				continue
			}
			instances := convertFargoAppToInstances(update.App)
			s.logger.Log("instances", len(instances))
			s.cache.Update(sd.Event{Instances: instances})

		case q := <-s.quitc:
			close(q)
			return
		}
	}
}

func (s *Instancer) getInstances() ([]string, error) {
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

// Register implements Instancer.
func (s *Instancer) Register(ch chan<- sd.Event) {
	s.cache.Register(ch)
}

// Deregister implements Instancer.
func (s *Instancer) Deregister(ch chan<- sd.Event) {
	s.cache.Deregister(ch)
}

// state returns the current state of instance.Cache, only for testing
func (s *Instancer) state() sd.Event {
	return s.cache.State()
}
