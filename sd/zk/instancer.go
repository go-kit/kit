package zk

import (
	"github.com/samuel/go-zookeeper/zk"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/internal/instance"
)

// Instancer yield instances stored in a certain ZooKeeper path. Any kind of
// change in that path is watched and will update the subscribers.
type Instancer struct {
	instance.Cache
	client Client
	path   string
	logger log.Logger
	quitc  chan struct{}
}

// NewInstancer returns a ZooKeeper Instancer. ZooKeeper will start watching
// the given path for changes and update the Instancer endpoints.
func NewInstancer(c Client, path string, logger log.Logger) (*Instancer, error) {
	s := &Instancer{
		Cache:  *instance.NewCache(),
		client: c,
		path:   path,
		logger: logger,
		quitc:  make(chan struct{}),
	}

	err := s.client.CreateParentNodes(s.path)
	if err != nil {
		return nil, err
	}

	instances, eventc, err := s.client.GetEntries(s.path)
	if err != nil {
		logger.Log("path", s.path, "msg", "failed to retrieve entries", "err", err)
		// TODO why zk constructor exits when other implementations continue?
		return nil, err
	}
	logger.Log("path", s.path, "instances", len(instances))
	s.Update(sd.Event{Instances: instances})

	go s.loop(eventc)

	return s, nil
}

func (s *Instancer) loop(eventc <-chan zk.Event) {
	var (
		instances []string
		err       error
	)
	for {
		select {
		case <-eventc:
			// We received a path update notification. Call GetEntries to
			// retrieve child node data, and set a new watch, as ZK watches are
			// one-time triggers.
			instances, eventc, err = s.client.GetEntries(s.path)
			if err != nil {
				s.logger.Log("path", s.path, "msg", "failed to retrieve entries", "err", err)
				s.Update(sd.Event{Err: err})
				continue
			}
			s.logger.Log("path", s.path, "instances", len(instances))
			s.Update(sd.Event{Instances: instances})

		case <-s.quitc:
			return
		}
	}
}

// Stop terminates the Instancer.
func (s *Instancer) Stop() {
	close(s.quitc)
}
