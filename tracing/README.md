# package tracing

`package tracing` provides [Dapper-style][dapper] request tracing to services.
An implementation exists for [Zipkin][]; [Appdash][] support is planned.

[dapper]: http://research.google.com/pubs/pub36356.html
[Zipkin]: https://blog.twitter.com/2012/distributed-systems-tracing-with-zipkin
[Appdash]: https://sourcegraph.com/blog/117580140734

## Rationale

TODO

## Usage

Wrap a [server.Endpoint][] so that it emits traces to a Zipkin collector.

[server.Endpoint]: http://godoc.org/github.com/go-kit/kit/server#Endpoint

```go
func main() {
	var (
		myHost        = "instance01.addsvc.internal.net"
		myMethod      = "ADD"
		scribeHost    = "scribe.internal.net"
		timeout       = 50 * time.Millisecond
		batchSize     = 100
		batchInterval = 3 * time.Second
	)

	spanFunc := zipkin.NewSpanFunc(myHost, myMethod)
	collector, _ := zipkin.NewScribeCollector(scribeHost, timeout, batchSize, batchInterval)

	var e server.Endpoint
	e = makeEndpoint() // for your service
	e = zipkin.AnnotateEndpoint(spanFunc, collector)

	serve(e)
}
```
