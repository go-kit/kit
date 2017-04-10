package eureka

import (
	"errors"
	"reflect"

	"github.com/go-kit/kit/log"
	"github.com/hudl/fargo"
)

type testConnection struct {
	instances      []*fargo.Instance
	application    *fargo.Application
	errInstances   error
	errApplication error
	errHeartbeat   error
	errRegister    error
	errDeregister  error
}

var (
	errTest       = errors.New("kaboom")
	errNotFound   = &fargoUnsuccessfulHTTPResponse{statusCode: 404, messagePrefix: "not found"}
	loggerTest    = log.NewNopLogger()
	appNameTest   = "go-kit"
	appUpdateTest = &fargo.Application{
		Name:      appNameTest,
		Instances: []*fargo.Instance{instanceTest1, instanceTest2},
	}
	instanceTest1 = &fargo.Instance{
		HostName:         "serveregistrar1.acme.org",
		Port:             8080,
		App:              appNameTest,
		IPAddr:           "192.168.0.1",
		VipAddress:       "192.168.0.1",
		SecureVipAddress: "192.168.0.1",
		HealthCheckUrl:   "http://serveregistrar1.acme.org:8080/healthz",
		StatusPageUrl:    "http://serveregistrar1.acme.org:8080/status",
		HomePageUrl:      "http://serveregistrar1.acme.org:8080/",
		Status:           fargo.UP,
		DataCenterInfo:   fargo.DataCenterInfo{Name: fargo.MyOwn},
		LeaseInfo:        fargo.LeaseInfo{RenewalIntervalInSecs: 1},
	}
	instanceTest2 = &fargo.Instance{
		HostName:         "serveregistrar2.acme.org",
		Port:             8080,
		App:              appNameTest,
		IPAddr:           "192.168.0.2",
		VipAddress:       "192.168.0.2",
		SecureVipAddress: "192.168.0.2",
		HealthCheckUrl:   "http://serveregistrar2.acme.org:8080/healthz",
		StatusPageUrl:    "http://serveregistrar2.acme.org:8080/status",
		HomePageUrl:      "http://serveregistrar2.acme.org:8080/",
		Status:           fargo.UP,
		DataCenterInfo:   fargo.DataCenterInfo{Name: fargo.MyOwn},
	}
)

var _ fargoConnection = (*testConnection)(nil)

func (c *testConnection) RegisterInstance(i *fargo.Instance) error {
	if c.errRegister == nil {
		for _, instance := range c.instances {
			if reflect.DeepEqual(*instance, *i) {
				return errors.New("already registered")
			}
		}

		c.instances = append(c.instances, i)
	}
	return c.errRegister
}

func (c *testConnection) HeartBeatInstance(i *fargo.Instance) error {
	return c.errHeartbeat
}

func (c *testConnection) DeregisterInstance(i *fargo.Instance) error {
	if c.errDeregister == nil {
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
	}
	return c.errDeregister
}

func (c *testConnection) ReregisterInstance(ins *fargo.Instance) error {
	return nil
}

func (c *testConnection) ScheduleAppUpdates(name string, await bool, done <-chan struct{}) <-chan fargo.AppUpdate {
	updatec := make(chan fargo.AppUpdate, 1)
	updatec <- fargo.AppUpdate{App: c.application, Err: c.errApplication}
	return updatec
}

func (c *testConnection) GetApp(name string) (*fargo.Application, error) {
	return &fargo.Application{Name: appNameTest, Instances: c.instances}, c.errInstances
}
