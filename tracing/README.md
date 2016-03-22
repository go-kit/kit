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

## Test Setup

Setting up [Zipkin] is not an easy thing to do. It will also demand quite some
resources. To help you get started with development and testing we've made a
[VirtualBox] image available through [Vagrant] (*The box will require about 6GB
internal memory*).

First make sure you've installed [Vagrant] on your machine. Then you can
download and run the [Vagrant] image like this from the command line:

```
# create a new directory to store your vagrant configuration and image files
mkdir zipkin
cd zipkin
vagrant init bas303/zipkin
vagrant up --provider virtualbox
```

[Zipkin]: http://zipkin.io/
[VirtualBox]: https://www.virtualbox.org/
[Vagrant]: https://www.vagrantup.com/

You probably need to adjust the `Vagrantfile` configuration to meet your
networking needs. The file itself is documented so should not be hard to get
configured. After the change you can reload your box with the updated settings
like this:

```
vagrant reload
```

As mentioned the box is quite heavy and may take a few minutes to fully boot up.
To get into the box connect through ssh and use `vagrant` for both username and
password.

The following services have been set-up to run:
- Apache ZooKeeper (port: 2181)
- Apache Kafka (port: 9092)
- MySQL Server 5.5 (port: 3306)
- Zipkin Collector (Kafka, MySQL)
- Zipkin Query (MySQL)
- Zipkin Web (port: 8080, 9990)

To inspect if everything booted up properly check the log files in these
directories:
```
/var/log/zookeeper
/var/log/kafka
/var/log/mysql
/var/log/zipkin
```

The individual services can be managed with the `service` command:
- zookeeper
- kafka
- mysql
- collector
- query
- web

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
