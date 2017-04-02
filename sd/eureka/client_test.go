package eureka

import (
	"errors"
	"reflect"

	fargo "github.com/hudl/fargo"

	"github.com/go-kit/kit/log"
)

var (
	errTest       = errors.New("kaboom")
	loggerTest    = log.NewNopLogger()
	instanceTest1 = &fargo.Instance{
		HostName:         "server1.acme.org",
		Port:             8080,
		App:              "go-kit",
		IPAddr:           "192.168.0.1",
		VipAddress:       "192.168.0.1",
		SecureVipAddress: "192.168.0.1",
		HealthCheckUrl:   "http://server1.acme.org:8080/healthz",
		StatusPageUrl:    "http://server1.acme.org:8080/status",
		HomePageUrl:      "http://server1.acme.org:8080/",
		Status:           fargo.UP,
		DataCenterInfo:   fargo.DataCenterInfo{Name: fargo.MyOwn},
		LeaseInfo:        fargo.LeaseInfo{RenewalIntervalInSecs: 1},
	}
	instanceTest2 = &fargo.Instance{
		HostName:         "server2.acme.org",
		Port:             8080,
		App:              "go-kit",
		IPAddr:           "192.168.0.2",
		VipAddress:       "192.168.0.2",
		SecureVipAddress: "192.168.0.2",
		HealthCheckUrl:   "http://server2.acme.org:8080/healthz",
		StatusPageUrl:    "http://server2.acme.org:8080/status",
		HomePageUrl:      "http://server2.acme.org:8080/",
		Status:           fargo.UP,
		DataCenterInfo:   fargo.DataCenterInfo{Name: fargo.MyOwn},
	}
	applicationTest = &fargo.Application{
		Name:      "go-kit",
		Instances: []*fargo.Instance{instanceTest1, instanceTest2},
	}
)

type testClient struct {
	instances      []*fargo.Instance
	application    *fargo.Application
	errInstances   error
	errApplication error
	errHeartbeat   error
}

func (c *testClient) Register(i *fargo.Instance) error {
	for _, instance := range c.instances {
		if reflect.DeepEqual(*instance, *i) {
			return errors.New("already registered")
		}
	}

	c.instances = append(c.instances, i)
	return nil
}

func (c *testClient) Deregister(i *fargo.Instance) error {
	var newInstances []*fargo.Instance
	for _, instance := range c.instances {
		if reflect.DeepEqual(*instance, *i) {
			continue
		}
		newInstances = append(newInstances, instance)
	}
	if len(newInstances) == len(c.instances) {
		return errors.New("not registered")
	}

	c.instances = newInstances
	return nil
}

func (c *testClient) Heartbeat(i *fargo.Instance) (err error) {
	return c.errHeartbeat
}

func (c *testClient) Instances(app string) ([]*fargo.Instance, error) {
	return c.instances, c.errInstances
}

func (c *testClient) ScheduleUpdates(service string, quitc chan struct{}) <-chan fargo.AppUpdate {
	updatec := make(chan fargo.AppUpdate, 1)
	updatec <- fargo.AppUpdate{App: c.application, Err: c.errApplication}
	return updatec
}
