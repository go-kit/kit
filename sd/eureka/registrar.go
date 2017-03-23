package eureka

import (
	"fmt"
	"sync"
	"time"

	stdeureka "github.com/hudl/fargo"

	"github.com/go-kit/kit/log"
)

// Registrar maintains service instance liveness information in Eureka.
type Registrar struct {
	client   Client
	instance *stdeureka.Instance
	logger   log.Logger

	quitmtx sync.Mutex
	quit    chan bool
}

// NewRegistrar returns an Eureka Registrar acting on behalf of the provided
// Fargo instance.
func NewRegistrar(client Client, i *stdeureka.Instance, l log.Logger) *Registrar {
	return &Registrar{
		client:   client,
		instance: i,
		logger:   log.With(l, "service", i.App, "address", fmt.Sprintf("%s:%d", i.IPAddr, i.Port)),
	}
}

// Register implements sd.Registrar interface.
func (r *Registrar) Register() {
	if err := r.client.Register(r.instance); err != nil {
		r.logger.Log("err", err)
	} else {
		r.logger.Log("action", "register")
	}

	if r.instance.LeaseInfo.RenewalIntervalInSecs > 0 {
		// User has opted for heartbeat functionality in Eureka.
		go r.loop()
	}
}

// Deregister implements sd.Registrar interface.
func (r *Registrar) Deregister() {
	if err := r.client.Deregister(r.instance); err != nil {
		r.logger.Log("err", err)
	} else {
		r.logger.Log("action", "deregister")
	}

	r.quitmtx.Lock()
	defer r.quitmtx.Unlock()
	if r.quit != nil {
		r.quit <- true
	}
}

func (r *Registrar) loop() {
	r.quitmtx.Lock()
	if r.quit != nil {
		defer r.quitmtx.Unlock()
		return // Already running.
	}
	r.quit = make(chan bool)
	r.quitmtx.Unlock()

	tick := time.NewTicker(time.Duration(r.instance.LeaseInfo.RenewalIntervalInSecs) * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			if err := r.client.Heartbeat(r.instance); err != nil {
				r.logger.Log("err", err)
			}
		case <-r.quit:
			r.quitmtx.Lock()
			defer r.quitmtx.Unlock()

			close(r.quit)
			r.quit = nil

			return
		}
	}
}
