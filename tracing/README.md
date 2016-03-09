# package tracing

`package tracing` provides [Dapper][]-style request tracing to services.
An implementation exists for [Zipkin][]; [Appdash][] support is planned.

[Dapper]: http://research.google.com/pubs/pub36356.html
[Zipkin]: https://blog.twitter.com/2012/distributed-systems-tracing-with-zipkin
[Appdash]: https://github.com/sourcegraph/appdash

## Rationale

Request tracing is a fundamental building block for large distributed
applications. It's instrumental in understanding request flows, identifying
hot spots, and diagnosing errors. All microservice infrastructures will
benefit from request tracing; sufficiently large infrastructures will require
it.

## Usage

Wrap a server- or client-side [endpoint][] so that it emits traces to a Zipkin
collector.

[endpoint]: http://godoc.org/github.com/go-kit/kit/endpoint#Endpoint

```go
func main() {
	var (
		myHost        = "instance01.addsvc.internal.net"
		myMethod      = "ADD"
		scribeHost    = "scribe.internal.net"
		timeout       = 50 * time.Millisecond
		batchSize     = 100
		batchInterval = 5 * time.Second
	)
	spanFunc := zipkin.NewSpanFunc(myHost, myMethod)
	collector, _ := zipkin.NewScribeCollector(scribeHost, timeout, batchSize, batchInterval)

	// Server-side
	var server endpoint.Endpoint
	server = makeEndpoint() // for your service
	server = zipkin.AnnotateServer(spanFunc, collector)(server)
	go serveViaHTTP(server)

	// Client-side
	before := httptransport.ClientBefore(zipkin.ToRequest(spanFunc))
	var client endpoint.Endpoint
	client = httptransport.NewClient(addr, codec, factory, before)
	client = zipkin.AnnotateClient(spanFunc, collector)(client)
}
```
