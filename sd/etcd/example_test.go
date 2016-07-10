package etcd

import (
	"fmt"
	"time"
	"io"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/endpoint"
	"golang.org/x/net/context"
)

// Package sd/etcd provides a wrapper around the coroes/etcd key value store (https://github.com/coreos/etcd)
// This example assumes the user has an instance of etcd installed and running locally on port 2379
func Example() {

	var (
		prefix   = "/services/foosvc/" // known at compile time
		instance = "1.2.3.4:8080"      // taken from runtime or platform, somehow
		key      = prefix + instance
		value    = "http://" + instance // based on our transport
	)

	client, err := NewClient(context.Background(), []string{"http://:2379"}, ClientOptions{
		DialTimeout:             2 * time.Second,
		DialKeepAlive:           2 * time.Second,
		HeaderTimeoutPerRequest: 2 * time.Second,
	})

	// Instantiate new instance of *Registrar passing in test data
	registrar := NewRegistrar(client, Service{
		Key:   key,
		Value: value,
	}, log.NewNopLogger())
	// Register new test data to etcd
	registrar.Register()

	//Retrieve entries from etcd
	_, err = client.GetEntries(key)
	if err != nil {
		fmt.Println(err)
	}

	factory := func(string) (endpoint.Endpoint, io.Closer, error) {
		return endpoint.Nop, nil, nil
	}
	subscriber, _ := NewSubscriber(client, prefix, factory, log.NewNopLogger())

	endpoints, err := subscriber.Endpoints()
	if err != nil {
		fmt.Printf("err: %v", err)
	}
	fmt.Println(len(endpoints)) // hopefully 1

	// Deregister first instance of test data
	registrar.Deregister()

	endpoints, err = subscriber.Endpoints()
	if err != nil {
		fmt.Printf("err: %v", err)
	}
	fmt.Println(len(endpoints)) // hopefully 0

	// Verify test data no longer exists in etcd
	_, err = client.GetEntries(key)
	if err != nil {
		fmt.Println(err)
	}
}
