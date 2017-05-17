// +build integration

package eureka

import (
	"io"
	"os"
	"testing"
	"time"

	"github.com/hudl/fargo"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
)

// Package sd/eureka provides a wrapper around the Netflix Eureka service
// registry by way of the Fargo library. This test assumes the user has an
// instance of Eureka available at the address in the environment variable.
// Example `${EUREKA_ADDR}` format: http://localhost:8761/eureka
//
// NOTE: when starting a Eureka server for integration testing, ensure
// the response cache interval is reduced to one second. This can be
// achieved with the following Java argument:
// `-Deureka.server.responseCacheUpdateIntervalMs=1000`
func TestIntegration(t *testing.T) {
	eurekaAddr := os.Getenv("EUREKA_ADDR")
	if eurekaAddr == "" {
		t.Skip("EUREKA_ADDR is not set")
	}

	logger := log.NewLogfmtLogger(os.Stderr)
	logger = log.With(logger, "ts", log.DefaultTimestamp)

	var fargoConfig fargo.Config
	// Target Eureka server(s).
	fargoConfig.Eureka.ServiceUrls = []string{eurekaAddr}
	// How often the subscriber should poll for updates.
	fargoConfig.Eureka.PollIntervalSeconds = 1

	// Create a Fargo connection and a Eureka registrar.
	fargoConnection := fargo.NewConnFromConfig(fargoConfig)
	registrar1 := NewRegistrar(&fargoConnection, instanceTest1, log.With(logger, "component", "registrar1"))

	// Register one instance.
	registrar1.Register()
	defer registrar1.Deregister()

	// This should be enough time for the Eureka server response cache to update.
	time.Sleep(time.Second)

	// Build a Eureka subscriber.
	factory := func(instance string) (endpoint.Endpoint, io.Closer, error) {
		t.Logf("factory invoked for %q", instance)
		return endpoint.Nop, nil, nil
	}
	instancer := NewInstancer(
		&fargoConnection,
		appNameTest,
		log.With(logger, "component", "instancer"),
	)
	defer instancer.Stop()
	endpointer := sd.NewEndpointer(instancer, factory, log.With(logger, "component", "endpointer"))

	// We should have one endpoint immediately after subscriber instantiation.
	endpoints, err := endpointer.Endpoints()
	if err != nil {
		t.Error(err)
	}
	if want, have := 1, len(endpoints); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	// Register a second instance
	registrar2 := NewRegistrar(&fargoConnection, instanceTest2, log.With(logger, "component", "registrar2"))
	registrar2.Register()
	defer registrar2.Deregister() // In case of exceptional circumstances.

	// This should be enough time for a scheduled update assuming Eureka is
	// configured with the properties mentioned in the function comments.
	time.Sleep(2 * time.Second)

	// Now we should have two endpoints.
	endpoints, err = endpointer.Endpoints()
	if err != nil {
		t.Error(err)
	}
	if want, have := 2, len(endpoints); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	// Deregister the second instance.
	registrar2.Deregister()

	// Wait for another scheduled update.
	time.Sleep(2 * time.Second)

	// And then there was one.
	endpoints, err = endpointer.Endpoints()
	if err != nil {
		t.Error(err)
	}
	if want, have := 1, len(endpoints); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}
