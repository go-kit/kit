package eureka

import (
	"github.com/hudl/fargo"
)

// Client is a wrapper around the Eureka API.
type Client interface {
	// Register an instance with Eureka.
	Register(i *fargo.Instance) error

	// Deregister an instance from Eureka.
	Deregister(i *fargo.Instance) error

	// Send an instance heartbeat to Eureka.
	Heartbeat(i *fargo.Instance) error

	// Get all instances for an app in Eureka.
	Instances(app string) ([]*fargo.Instance, error)

	// Receive scheduled updates about an app's instances in Eureka.
	ScheduleUpdates(app string, quitc chan struct{}) <-chan fargo.AppUpdate
}

type client struct {
	connection *fargo.EurekaConnection
}

// NewClient returns an implementation of the Client interface, wrapping a
// concrete connection to Eureka using the Fargo library.
// Taking in Fargo's own connection abstraction gives the user maximum
// freedom in regards to how that connection is configured.
func NewClient(ec *fargo.EurekaConnection) Client {
	return &client{connection: ec}
}

func (c *client) Register(i *fargo.Instance) error {
	if c.instanceRegistered(i) {
		// Already registered. Send a heartbeat instead.
		return c.Heartbeat(i)
	}
	return c.connection.RegisterInstance(i)
}

func (c *client) Deregister(i *fargo.Instance) error {
	return c.connection.DeregisterInstance(i)
}

func (c *client) Heartbeat(i *fargo.Instance) (err error) {
	if err = c.connection.HeartBeatInstance(i); err != nil && c.instanceNotFoundErr(err) {
		// Instance not registered. Register first before sending heartbeats.
		return c.Register(i)
	}
	return err
}

func (c *client) Instances(app string) ([]*fargo.Instance, error) {
	stdApp, err := c.connection.GetApp(app)
	if err != nil {
		return nil, err
	}
	return stdApp.Instances, nil
}

func (c *client) ScheduleUpdates(app string, quitc chan struct{}) <-chan fargo.AppUpdate {
	return c.connection.ScheduleAppUpdates(app, false, quitc)
}

func (c *client) instanceRegistered(i *fargo.Instance) bool {
	_, err := c.connection.GetInstance(i.App, i.Id())
	return err == nil
}

func (c *client) instanceNotFoundErr(err error) bool {
	code, ok := fargo.HTTPResponseStatusCode(err)
	return ok && code == 404
}
