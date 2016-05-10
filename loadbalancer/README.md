# package loadbalancer

`package loadbalancer` provides a client-side load balancer abstraction.

A publisher is responsible for emitting the most recent set of endpoints for a
single logical service. Publishers exist for static endpoints, and endpoints
discovered via periodic DNS SRV lookups on a single logical name. Consul and
etcd publishers are planned.

Different load balancers are implemented on top of publishers. Go kit
currently provides random and round-robin load balancers. Smarter behaviors,
e.g. load balancing based on underlying endpoint priority/weight, is planned.

## Rationale

TODO

## Usage

In your client, construct a publisher for a specific remote service, and pass
it to a load balancer. Then, request an endpoint from the load balancer
whenever you need to make a request to that remote service.

```go
import (
	"github.com/go-kit/kit/loadbalancer"
	"github.com/go-kit/kit/loadbalancer/dnssrv"
)

func main() {
	// Construct a load balancer for foosvc, which gets foosvc instances by
	// polling a specific DNS SRV name.
	p, err := dnssrv.NewPublisher("foosvc.internal.domain", 5*time.Second, fooFactory, logger)
	if err != nil {
		panic(err)
	}
	
	lb := loadbalancer.NewRoundRobin(p)

	// Get a new endpoint from the load balancer.
	endpoint, err := lb.Endpoint()
	if err != nil {
		panic(err)
	}

	// Use the endpoint to make a request.
	response, err := endpoint(ctx, request)
}

func fooFactory(instance string) (endpoint.Endpoint, error) {
	// Convert an instance (host:port) to an endpoint, via a defined transport binding.
}
```

It's also possible to wrap a load balancer with a retry strategy, so that it
can be used as an endpoint directly. This may make load balancers more
convenient to use, at the cost of fine-grained control of failures.

```go
func main() {
	p := dnssrv.NewPublisher("foosvc.internal.domain", 5*time.Second, fooFactory, logger)
	lb := loadbalancer.NewRoundRobin(p)
	endpoint := loadbalancer.Retry(3, 5*time.Seconds, lb)

	response, err := endpoint(ctx, request) // requests will be automatically load balanced
}
```
