# Zipkin

## Development and Testing Set-up

Setting up [Zipkin] is not an easy thing to do. It will also demand quite some
resources. To help you get started with development and testing we've made a
docker-compose file available for running a full Zipkin stack.

You will need [docker-compose] 1.6.0+ and [docker-engine] 1.10.0+.

If running on Linux `HOSTNAME` can be set to `localhost`. If running on Mac OS X
or Windows you probably need to set the hostname environment variable to the
hostname of the VM running the docker containers.

```sh
cd tracing/zipkin
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

Follow the [addsvc] example to check out how to wire the Zipkin Middleware. The
changes should be relatively minor.

The [zipkin-go-opentracing] package has support for Kafka and Scribe collectors
as well as using Go Kit's [Log] package for logging.

### Span per Node vs. Span per RPC
OpenTracing tends to support the `span per node` model, meaning the server and
client side of an RPC call are dealt with in separate spans. This diverts from
the Zipkin V1 model which is more tailored towards the `span per RPC call` model
in which the server and client side annotations of the same RPC call are part of
the same span. The [zipkin-go-opentracing] implementation allows you to choose
which model you wish to use. Make sure you select the same model consistently
for all your services that are required to talk to each other or you will have
trace propagation issues. If using non OpenTracing / legacy instrumentation use
the `span per RPC call` model.

To adhere to the OpenTracing philosophy the Tracer defaults to `span per node`.
To set the `span per RPC call` mode start your tracer like this:

```go
tracer, err = zipkin.NewTracer(
	zipkin.NewRecorder(...),
	zipkin.ClientServerSameSpan(true),
)
```

[zipkin-go-opentracing]: https://github.com/openzipkin/zipkin-go-opentracing
[addsvc]:https://github.com/go-kit/kit/tree/master/examples/addsvc
[Log]: https://github.com/go-kit/kit/tree/master/log

### Tracing Resources

In our legacy implementation we had the `NewChildSpan` method to allow
annotation of resources such as databases, caches and other services that do not
have server side tracing support. Since OpenTracing has no specific method of
dealing with these items explicitely that is compatible with Zipkin's `SA`
annotation the [zipkin-go-opentracing] has implemented support using the
OpenTracing Tags system. Here is an example of how one would be able to record
a resource span compatible with standard OpenTracing and triggering an `SA`
annotation in [zipkin-go-opentracing]:

```go
// you need to import the ext package for the Tag helper functions
import (
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func (svc *Service) GetMeSomeExamples(ctx context.Context, ...) ([]Examples, error) {
	// Example of annotating a database query:
	var (
		serviceName = "MySQL"
		serviceHost = "mysql.example.com"
		servicePort = uint16(3306)
		queryLabel  = "GetExamplesByParam"
		query       = "select * from example where param = 'value'"
	)

	// retrieve the parent span, if not found create a new trace
	parentSpan := opentracing.SpanFromContext(ctx)
	if parentSpan == nil {
		parentSpan = opentracing.StartSpan(queryLabel)
	}

	// create a new span to record the resource interaction
	span := opentracing.StartChildSpan(parentSpan, queryLabel)

	// span.kind "resource" triggers SA annotation
	ext.SpanKind.Set(span, "resource")

	// this will label the span's service & hostPort (called Endpoint in Zipkin)
	ext.PeerService.Set(span, serviceName)
	ext.PeerHostname,Set(span, serviceHost)
	ext.PeerPort.Set(span, servicePort)

	// a Tag is the equivalent of a Zipkin Binary Annotation (key:value pair)
	span.SetTag("query", query)

	// a LogEvent is the equivalent of a Zipkin Annotation (timestamped)
	span.LogEvent("query:start")

	// do the actual query...

	// let's annotate the end...
	span.LogEvent("query:end")

	// we're done with the span. send it to the SpanRecorder for collection...
	span.Finish()

	// do other stuff
	...
}
```
