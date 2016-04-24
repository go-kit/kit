# Zipkin

## Development and Testing Set-up

Setting up [Zipkin] is not an easy thing to do. It will also demand quite some
resources. To help you get started with development and testing we've made a
docker-compose file available for running a full Zipkin stack. See the
`kit/tracing/zipkin/_docker` subdirectory.

You will need [docker-compose] 1.6.0+ and [docker-engine] 1.10.0+.

If running on Linux `HOSTNAME` can be set to `localhost`. If running on Mac OS X
or Windows you probably need to set the hostname environment variable to the
hostname of the VM running the docker containers.

```sh
cd tracing/zipkin/_docker
HOSTNAME=localhost docker-compose -f docker-compose-zipkin.yml up
```

[Zipkin]: http://zipkin.io/
[docker-compose]: https://docs.docker.com/compose/
[docker-engine]: https://docs.docker.com/engine/

As mentioned the [Zipkin] stack is quite heavy and may take a few minutes to
fully initialize.

The following services have been set-up to run:
- Apache Cassandra (port: 9160 (thrift), 9042 (native))
- Apache ZooKeeper (port: 2181)
- Apache Kafka (port: 9092)
- Zipkin Collector
- Zipkin Query
- Zipkin Web (port: 8080, 9990)


## Middleware Usage

Wrap a server- or client-side [endpoint][] so that it emits traces to a Zipkin
collector. Make sure the host given to `MakeNewSpanFunc` resolves to an IP. If
not your span will silently fail!

[endpoint]: http://godoc.org/github.com/go-kit/kit/endpoint#Endpoint

If needing to create child spans in methods or calling another service from your
service method, it is highly recommended to request a context parameter so you
can transfer the needed metadata for traces across service boundaries.

It is also wise to always return error parameters with your service method
calls, even if your service method implementations will not throw errors
themselves. The error return parameter can be wired to pass the potential
transport errors when consuming your service API in a networked environment.

```go
func main() {
	var (
		// myHost MUST resolve to an IP or your span will not show up in Zipkin.
		myHost        = "instance01.addsvc.internal.net:8000"
		myService     = "AddService"
		myMethod      = "Add"
		url           = myHost + "/add/"
		kafkaHost     = []string{"kafka.internal.net:9092"}
	)

	ctx := context.Background()

	// Set Up Zipkin Collector and Span factory
	spanFunc := zipkin.MakeNewSpanFunc(myHost, myService, myMethod)
	collector, _ := zipkin.NewKafkaCollector(kafkaHost)

	// Server-side Wiring
	var server endpoint.Endpoint
	server = makeEndpoint() // for your service
	// wrap endpoint with Zipkin tracing middleware
	server = zipkin.AnnotateServer(spanFunc, collector)(server)

	http.Handle(
		"/add/",
		httptransport.NewServer(
			ctx,
			server,
			decodeRequestFunc,
			encodeResponseFunc,
			httptransport.ServerBefore(
				zipkin.ToContext(spanFunc),
			),
		),
	)
	...

	// Client-side
	var client endpoint.Endpoint
	client = httptransport.NewClient(
		"GET",
		URL,
		encodeRequestFunc,
		decodeResponseFunc,
		httptransport.ClientBefore(zipkin.ToRequest(spanFunc)),
	).Endpoint()
	client = zipkin.AnnotateClient(spanFunc, collector)(client)

	ctx, cancel := context.WithTimeout(ctx, myTimeout)
	defer cancel()

	reply, err := client(ctx, param1, param2)
	// do something with the response/error
	...
}
```

## Annotating Remote Resources

Next to the above shown examples of wiring server-side and client-side tracing
middlewares, you can also span resources called from your service methods.

To do this, the service method needs to include a context parameter. From your
endpoint wrapper you can inject the endpoint context which will hold the parent
span already created by the server-side middleware. If the resource is a remote
database you can use the `zipkin.ServerAddr` spanOption to identify the remote
host:port and the display name of this resource.

```go
type MyService struct {
	// add a Zipkin Collector to your service implementation's properties.
	Collector zipkin.Collector
}

// Example of the endpoint.Endpoint to service method wrapper, injecting the
// context provided by the transport server.
func makeComplexEndpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ComplexRequest)
		v, err := svc.Complex(ctx, req.A, req.B)
		return ComplexResponse{V: v, Err: err}, nil
	}
}

// Complex is an example method of our service, displaying the tracing of a
// remote database resource.
func (s *MyService) Complex(ctx context.Context, A someType, B otherType) (returnType, error) {
	// we've parsed the incoming parameters and now we need to query the database.
	// we wish to include this action into our trace.
	span, collect := zipkin.NewChildSpan(
		ctx,
		s.Collector,
		"complexQuery",
		zipkin.ServerAddr(
			"mysql01.internal.net:3306",
			"MySQL",
		),
	)
	// you probably want to binary annotate your query
	span.AnnotateBinary("query", "SELECT ... FROM ... WHERE ... ORDER BY ..."),
	// annotate the start of the query
	span.Annotate("complexQuery:start")
	// do the query and handle resultset
	...
	// annotate we are done with the query
	span.Annotate("complexQuery:end")
	// maybe binary annotate some items returned by the resultset
	...
	// when done with all annotations, collect the span
	collect()
	...
}
```
