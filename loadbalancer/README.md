# package loadbalancer

`package loadbalancer` provides a client-side load balancer abstraction.

A publisher is responsible for emitting the most recent set of endpoints for a
single logical service. Publishers exist for static endpoints, and endpoints
discovered via periodic DNS SRV lookups on a single logical name. Consul and
etcd publishers are planned.

Different load balancing strategies are implemented on top of publishers. Go
kit currently provides random and round-robin semantics. Smarter behaviors,
e.g. load balancing based on underlying endpoint priority/weight, is planned.

## Rationale

TODO

## Usage

In your client, define a publisher, wrap it with a balancing strategy, and pass
it to a retry strategy, which returns an endpoint.  Use that endpoint to make
requests, or wrap it with other value-add middleware.

```go
func main() {
	var (
		fooPublisher = loadbalancer.NewDNSSRVPublisher("foo.mynet.local", 5*time.Second, makeEndpoint)
		fooBalancer  = loadbalancer.RoundRobin(mysvcPublisher)
		fooEndpoint  = loadbalancer.Retry(3, time.Second, fooBalancer)
	)
	http.HandleFunc("/", handle(fooEndpoint))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func makeEndpoint(hostport string) endpoint.Endpoint {
	// Convert a host:port to a endpoint via your defined transport.
}

func handle(foo endpoint.Endpoint) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// foo is usable as a load-balanced remote endpoint.
	}
}
```
