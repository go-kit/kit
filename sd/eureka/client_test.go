package eureka

import (
	"errors"
	"reflect"

	"github.com/go-kit/kit/log"
	stdeureka "github.com/hudl/fargo"
)

var (
	errTest       = errors.New("kaboom")
	loggerTest    = log.NewNopLogger()
	instanceTest1 = &stdeureka.Instance{
		HostName:         "server1.acme.org",
		Port:             8080,
		App:              "go-kit",
		IPAddr:           "192.168.0.1",
		VipAddress:       "192.168.0.1",
		SecureVipAddress: "192.168.0.1",
		HealthCheckUrl:   "http://server1.acme.org:8080/healthz",
		StatusPageUrl:    "http://server1.acme.org:8080/status",
		HomePageUrl:      "http://server1.acme.org:8080/",
		Status:           stdeureka.UP,
		DataCenterInfo:   stdeureka.DataCenterInfo{Name: stdeureka.MyOwn},
		LeaseInfo:        stdeureka.LeaseInfo{RenewalIntervalInSecs: 1},
	}
	instanceTest2 = &stdeureka.Instance{
		HostName:         "server2.acme.org",
		Port:             8080,
		App:              "go-kit",
		IPAddr:           "192.168.0.2",
		VipAddress:       "192.168.0.2",
		SecureVipAddress: "192.168.0.2",
		HealthCheckUrl:   "http://server2.acme.org:8080/healthz",
		StatusPageUrl:    "http://server2.acme.org:8080/status",
		HomePageUrl:      "http://server2.acme.org:8080/",
		Status:           stdeureka.UP,
		DataCenterInfo:   stdeureka.DataCenterInfo{Name: stdeureka.MyOwn},
	}
	applicationTest = &stdeureka.Application{
		Name:      "go-kit",
		Instances: []*stdeureka.Instance{instanceTest1, instanceTest2},
	}
)

type testClient struct {
	instances      []*stdeureka.Instance
	application    *stdeureka.Application
	errInstances   error
	errApplication error
	errHeartbeat   error
}

func (c *testClient) Register(i *stdeureka.Instance) error {
	for _, instance := range c.instances {
		if reflect.DeepEqual(*instance, *i) {
			return errors.New("already registered")
		}
	}

	c.instances = append(c.instances, i)
	return nil
}

func (c *testClient) Deregister(i *stdeureka.Instance) error {
	var newInstances []*stdeureka.Instance
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

func (c *testClient) Heartbeat(i *stdeureka.Instance) (err error) {
	return c.errHeartbeat
}

func (c *testClient) Instances(app string) ([]*stdeureka.Instance, error) {
	return c.instances, c.errInstances
}

func (c *testClient) ScheduleUpdates(service string, quitc chan struct{}) <-chan stdeureka.AppUpdate {
	updatec := make(chan stdeureka.AppUpdate, 1)
	updatec <- stdeureka.AppUpdate{App: c.application, Err: c.errApplication}
	return updatec
}
