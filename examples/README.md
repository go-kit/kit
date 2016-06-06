# Examples

1. [A minimal example](#a-minimal-example)
	1. [Your business logic](#your-business-logic)
	1. [Requests and responses](#requests-and-responses)
	1. [Endpoints](#endpoints)
	1. [Transports](#transports)
	1. [stringsvc1](#stringsvc1)
1. [Logging and instrumentation](#logging-and-instrumentation)
	1. [Transport logging](#transport-logging)
	1. [Application logging](#application-logging)
	1. [Instrumentation](#instrumentation)
	1. [stringsvc2](#stringsvc2)
1. [Calling other services](#calling-other-services)
	1. [Client-side endpoints](#client-side-endpoints)
	1. [Service discovery and load balancing](#service-discovery-and-load-balancing)
	1. [stringsvc3](#stringsvc3)
1. [Advanced topics](#advanced-topics)
	1. [Creating a client package](#creating-a-client-package)
	1. [Request tracing](#request-tracing)
	1. [Threading a context](#threading-a-context)
1. [Other examples](#other-examples)
	1. [Transport-specific](#transport-specific)
	1. [addsvc](#addsvc)
	1. [apigateway](#apigateway)

## A minimal example

Let's create a minimal Go kit service.

### Your business logic

Your service starts with your business logic.
In Go kit, we model a service as an **interface**.

```go
// StringService provides operations on strings.
type StringService interface {
	Uppercase(string) (string, error)
	Count(string) int
}
```

That interface will have an implementation.

```go
type stringService struct{}

func (stringService) Uppercase(s string) (string, error) {
	if s == "" {
		return "", ErrEmpty
	}
	return strings.ToUpper(s), nil
}

func (stringService) Count(s string) int {
	return len(s)
}

// ErrEmpty is returned when input string is empty
var ErrEmpty = errors.New("Empty string")
```

### Requests and responses

In Go kit, the primary messaging pattern is RPC.
So, every method in our interface will be modeled as a remote procedure call.
For each method, we define **request and response** structs,
 capturing all of the input and output parameters respectively.

```go
type uppercaseRequest struct {
	S string `json:"s"`
}

type uppercaseResponse struct {
	V   string `json:"v"`
	Err string `json:"err,omitempty"` // errors don't JSON-marshal, so we use a string
}

type countRequest struct {
	S string `json:"s"`
}

type countResponse struct {
	V int `json:"v"`
}
```

### Endpoints

Go kit provides much of its functionality through an abstraction called an **endpoint**.

```go
type Endpoint func(ctx context.Context, request interface{}) (response interface{}, err error)
```

An endpoint represents a single RPC.
That is, a single method in our service interface.
We'll write simple adapters to convert each of our service's methods into an endpoint.
Each adapter takes a StringService, and returns an endpoint that corresponds to one of the methods.

```go
import (
	"golang.org/x/net/context"
	"github.com/go-kit/kit/endpoint"
)

func makeUppercaseEndpoint(svc StringService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(uppercaseRequest)
		v, err := svc.Uppercase(req.S)
		if err != nil {
			return uppercaseResponse{v, err.Error()}, nil
		}
		return uppercaseResponse{v, ""}, nil
	}
}

func makeCountEndpoint(svc StringService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(countRequest)
		v := svc.Count(req.S)
		return countResponse{v}, nil
	}
}
```

### Transports

Now we need to expose your service to the outside world, so it can be called.
Your organization probably already has opinions about how services should talk to each other.
Maybe you use Thrift, or custom JSON over HTTP.
Go kit supports many **transports** out of the box.
(Adding support for new ones is easy—just [file an issue](https://github.com/go-kit/kit/issues).)

For this minimal example service, let's use JSON over HTTP.
Go kit provides a helper struct, in package transport/http.

```go
import (
	"encoding/json"
	"log"
	"net/http"

	"golang.org/x/net/context"

	httptransport "github.com/go-kit/kit/transport/http"
)

func main() {
	ctx := context.Background()
	svc := stringService{}

	uppercaseHandler := httptransport.NewServer(
		ctx,
		makeUppercaseEndpoint(svc),
		decodeUppercaseRequest,
		encodeResponse,
	)

	countHandler := httptransport.NewServer(
		ctx,
		makeCountEndpoint(svc),
		decodeCountRequest,
		encodeResponse,
	)

	http.Handle("/uppercase", uppercaseHandler)
	http.Handle("/count", countHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func decodeUppercaseRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request uppercaseRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func decodeCountRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request countRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}
```

### stringsvc1

The complete service so far is [stringsvc1][].

[stringsvc1]: https://github.com/go-kit/kit/blob/master/examples/stringsvc1

```
$ go get github.com/go-kit/kit/examples/stringsvc1
$ stringsvc1
```

```
$ curl -XPOST -d'{"s":"hello, world"}' localhost:8080/uppercase
{"v":"HELLO, WORLD","err":null}
$ curl -XPOST -d'{"s":"hello, world"}' localhost:8080/count
{"v":12}
```

## Logging and instrumentation

No service can be considered production-ready without thorough logging and instrumentation.

### Transport logging

Any component that needs to log should treat the logger like a dependency, same as a database connection.
So, we construct our logger in our `func main`, and pass it to components that need it.
We never use a globally-scoped logger.

We could pass a logger directly into our stringService implementation, but there's a better way.
Let's use a **middleware**, also known as a decorator.
A middleware is a function that takes an endpoint and returns an endpoint.

```go
type Middleware func(Endpoint) Endpoint
```

In between, it can do anything.
Let's create a basic logging middleware.

```go
func loggingMiddleware(logger log.Logger) Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			logger.Log("msg", "calling endpoint")
			defer logger.Log("msg", "called endpoint")
			return next(ctx, request)
		}
	}
}
```

And wire it into each of our handlers.

```go
logger := log.NewLogfmtLogger(os.Stderr)

svc := stringService{}

var uppercase endpoint.Endpoint
uppercase = makeUppercaseEndpoint(svc)
uppercase = loggingMiddleware(log.NewContext(logger).With("method", "uppercase"))(uppercase)

var count endpoint.Endpoint
count = makeCountEndpoint(svc)
count = loggingMiddleware(log.NewContext(logger).With("method", "count"))(count)

uppercaseHandler := httptransport.Server(
	// ...
	uppercase,
	// ...
)

countHandler := httptransport.Server(
	// ...
	count,
	// ...
)
```

It turns out that this technique is useful for a lot more than just logging.
Many Go kit components are implemented as endpoint middlewares.

### Application logging

But what if we want to log in our application domain, like the parameters that are passed in?
It turns out that we can define a middleware for our service, and get the same nice and composable effects.
Since our StringService is defined as an interface, we just need to make a new type
 which wraps an existing StringService, and performs the extra logging duties.

```go
type loggingMiddleware struct {
	logger log.Logger
	next   StringService
}

func (mw loggingMiddleware) Uppercase(s string) (output string, err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "uppercase",
			"input", s,
			"output", output,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	output, err = mw.next.Uppercase(s)
	return
}

func (mw loggingMiddleware) Count(s string) (n int) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "count",
			"input", s,
			"n", n,
			"took", time.Since(begin),
		)
	}(time.Now())

	n = mw.next.Count(s)
	return
}
```

And wire it in.

```go
import (
	"os"

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
)

func main() {
	logger := log.NewLogfmtLogger(os.Stderr)

	var svc StringService
	svc = stringsvc{}
	svc = loggingMiddleware{logger, svc}

	// ...

	uppercaseHandler := httptransport.NewServer(
		// ...
		makeUppercaseEndpoint(svc),
		// ...
	)

	countHandler := httptransport.NewServer(
		// ...
		makeCountEndpoint(svc),
		// ...
	)
}
```

Use endpoint middlewares for transport-domain concerns, like circuit breaking and rate limiting.
Use service middlewares for business-domain concerns, like logging and instrumentation.
Speaking of instrumentation...

### Instrumentation

In Go kit, instrumentation means using **package metrics** to record statistics about your service's runtime behavior.
Counting the number of jobs processed,
 recording the duration of requests after they've finished,
  and tracking the number of in-flight operations would all be considered instrumentation.

We can use the same middleware pattern that we used for logging.

```go
type instrumentingMiddleware struct {
	requestCount   metrics.Counter
	requestLatency metrics.TimeHistogram
	countResult    metrics.Histogram
	next           StringService
}

func (mw instrumentingMiddleware) Uppercase(s string) (output string, err error) {
	defer func(begin time.Time) {
		methodField := metrics.Field{Key: "method", Value: "uppercase"}
		errorField := metrics.Field{Key: "error", Value: fmt.Sprintf("%v", err)}
		mw.requestCount.With(methodField).With(errorField).Add(1)
		mw.requestLatency.With(methodField).With(errorField).Observe(time.Since(begin))
	}(time.Now())

	output, err = mw.next.Uppercase(s)
	return
}

func (mw instrumentingMiddleware) Count(s string) (n int) {
	defer func(begin time.Time) {
		methodField := metrics.Field{Key: "method", Value: "count"}
		errorField := metrics.Field{Key: "error", Value: fmt.Sprintf("%v", error(nil))}
		mw.requestCount.With(methodField).With(errorField).Add(1)
		mw.requestLatency.With(methodField).With(errorField).Observe(time.Since(begin))
		mw.countResult.Observe(int64(n))
	}(time.Now())

	n = mw.next.Count(s)
	return
}
```

And wire it into our service.

```go
import (
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/go-kit/kit/metrics"
)

func main() {
	logger := log.NewLogfmtLogger(os.Stderr)

	fieldKeys := []string{"method", "error"}
	requestCount := kitprometheus.NewCounter(stdprometheus.CounterOpts{
		// ...
	}, fieldKeys)
	requestLatency := metrics.NewTimeHistogram(time.Microsecond, kitprometheus.NewSummary(stdprometheus.SummaryOpts{
		// ...
	}, fieldKeys))
	countResult := kitprometheus.NewSummary(stdprometheus.SummaryOpts{
		// ...
	}, []string{}))

	var svc StringService
	svc = stringService{}
	svc = loggingMiddleware{logger, svc}
	svc = instrumentingMiddleware{requestCount, requestLatency, countResult, svc}

	// ...

	http.Handle("/metrics", stdprometheus.Handler())
}
```

### stringsvc2

The complete service so far is [stringsvc2][].

[stringsvc2]: https://github.com/go-kit/kit/blob/master/examples/stringsvc2

```
$ go get github.com/go-kit/kit/examples/stringsvc2
$ stringsvc2
msg=HTTP addr=:8080
```

```
$ curl -XPOST -d'{"s":"hello, world"}' localhost:8080/uppercase
{"v":"HELLO, WORLD","err":null}
$ curl -XPOST -d'{"s":"hello, world"}' localhost:8080/count
{"v":12}
```

```
method=uppercase input="hello, world" output="HELLO, WORLD" err=null took=2.455µs
method=count input="hello, world" n=12 took=743ns
```

## Calling other services

It's rare that a service exists in a vacuum.
Often, you need to call other services.
**This is where Go kit shines**.
We provide transport middlewares to solve many of the problems that come up.

Let's say that we want to have our string service call out to a _different_ string service
 to satisfy the Uppercase method.
In effect, proxying the request to another service.
Let's implement the proxying middleware as a ServiceMiddleware, same as a logging or instrumenting middleware.

```go
// proxymw implements StringService, forwarding Uppercase requests to the
// provided endpoint, and serving all other (i.e. Count) requests via the
// next StringService.
type proxymw struct {
	ctx       context.Context
	next      StringService     // Serve most requests via this service...
	uppercase endpoint.Endpoint // ...except Uppercase, which gets served by this endpoint
}
```

### Client-side endpoints

We've got exactly the same endpoint we already know about, but we'll use it to invoke, rather than serve, a request.
When used this way, we call it a _client_ endpoint.
And to invoke the client endpoint, we just do some simple conversions.

```go
func (mw proxymw) Uppercase(s string) (string, error) {
	response, err := mw.uppercase(mw.Context, uppercaseRequest{S: s})
	if err != nil {
		return "", err
	}
	resp := response.(uppercaseResponse)
	if resp.Err != "" {
		return resp.V, errors.New(resp.Err)
	}
	return resp.V, nil
}
```

Now, to construct one of these proxying middlewares, we convert a proxy URL string to an endpoint.
If we assume JSON over HTTP, we can use a helper in the transport/http package.

```go
import (
	httptransport "github.com/go-kit/kit/transport/http"
)

func proxyingMiddleware(proxyURL string, ctx context.Context) ServiceMiddleware {
	return func(next StringService) StringService {
		return proxymw{ctx, next, makeUppercaseEndpoint(ctx, proxyURL)}
	}
}

func makeUppercaseEndpoint(ctx context.Context, proxyURL string) endpoint.Endpoint {
	return httptransport.NewClient(
		"GET",
		mustParseURL(proxyURL),
		encodeUppercaseRequest,
		decodeUppercaseResponse,
	).Endpoint()
}
```

### Service discovery and load balancing

That's fine if we only have a single remote service.
But in reality, we'll probably have many service instances available to us.
We want to discover them through some service discovery mechanism, and spread our load across all of them.
And if any of those instances start to behave badly, we want to deal with that, without affecting our own service's reliability.

Go kit offers adapters to different service discovery systems, to get up-to-date sets of instances, exposed as individual endpoints.
Those adapters are called subscribers.

```go
type Subscriber interface {
	Endpoints() ([]endpoint.Endpoint, error)
}
```

Internally, subscribers use a provided factory function to convert each discovered instance string (typically host:port) to a usable endpoint.

```go
type Factory func(instance string) (endpoint.Endpoint, error)
```

So far, our factory function, makeUppercaseEndpoint, just calls the URL directly.
But it's important to put some safety middleware, like circuit breakers and rate limiters, into your factory, too.

```go
var e endpoint.Endpoint
e = makeUppercaseProxy(ctx, instance)
e = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(e)
e = kitratelimit.NewTokenBucketLimiter(jujuratelimit.NewBucketWithRate(float64(maxQPS), int64(maxQPS)))(e)
}
```

Now that we've got a set of endpoints, we need to choose one.
Load balancers wrap subscribers, and select one endpoint from many.
Go kit provides a couple of basic load balancers, and it's easy to write your own if you want more advanced heuristics.

```go
type Balancer interface {
	Endpoint() (endpoint.Endpoint, error)
}
```

Now we have the ability to choose endpoints according to some heuristic.
We can use that to provide a single, logical, robust endpoint to consumers.
A retry strategy wraps a load balancer, and returns a usable endpoint.
The retry strategy will retry failed requests until either the max attempts or timeout has been reached.

```go
func Retry(max int, timeout time.Duration, lb Balancer) endpoint.Endpoint
```

Let's wire up our final proxying middleware.
For simplicity, we'll assume the user will specify multiple comma-separate instance endpoints with a flag.

```go
func proxyingMiddleware(instances string, ctx context.Context, logger log.Logger) ServiceMiddleware {
	// If instances is empty, don't proxy.
	if instances == "" {
		logger.Log("proxy_to", "none")
		return func(next StringService) StringService { return next }
	}

	// Set some parameters for our client.
	var (
		qps         = 100                    // beyond which we will return an error
		maxAttempts = 3                      // per request, before giving up
		maxTime     = 250 * time.Millisecond // wallclock time, before giving up
	)

	// Otherwise, construct an endpoint for each instance in the list, and add
	// it to a fixed set of endpoints. In a real service, rather than doing this
	// by hand, you'd probably use package sd's support for your service
	// discovery system.
	var (
		instanceList = split(instances)
		subscriber   sd.FixedSubscriber
	)
	logger.Log("proxy_to", fmt.Sprint(instanceList))
	for _, instance := range instanceList {
		var e endpoint.Endpoint
		e = makeUppercaseProxy(ctx, instance)
		e = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(e)
		e = kitratelimit.NewTokenBucketLimiter(jujuratelimit.NewBucketWithRate(float64(qps), int64(qps)))(e)
		subscriber = append(subscriber, e)
	}

	// Now, build a single, retrying, load-balancing endpoint out of all of
	// those individual endpoints.
	balancer := lb.NewRoundRobin(subscriber)
	retry := lb.Retry(maxAttempts, maxTime, balancer)

	// And finally, return the ServiceMiddleware, implemented by proxymw.
	return func(next StringService) StringService {
		return proxymw{ctx, next, retry}
	}
}
```

### stringsvc3

The complete service so far is [stringsvc3][].

[stringsvc3]: https://github.com/go-kit/kit/blob/master/examples/stringsvc3

```
$ go get github.com/go-kit/kit/examples/stringsvc3
$ stringsvc3 -listen=:8001 &
listen=:8001 caller=proxying.go:25 proxy_to=none
listen=:8001 caller=main.go:72 msg=HTTP addr=:8001
$ stringsvc3 -listen=:8002 &
listen=:8002 caller=proxying.go:25 proxy_to=none
listen=:8002 caller=main.go:72 msg=HTTP addr=:8002
$ stringsvc3 -listen=:8003 &
listen=:8003 caller=proxying.go:25 proxy_to=none
listen=:8003 caller=main.go:72 msg=HTTP addr=:8003
$ stringsvc3 -listen=:8080 -proxy=localhost:8001,localhost:8002,localhost:8003
listen=:8080 caller=proxying.go:29 proxy_to="[localhost:8001 localhost:8002 localhost:8003]"
listen=:8080 caller=main.go:72 msg=HTTP addr=:8080
```

```
$ for s in foo bar baz ; do curl -d"{\"s\":\"$s\"}" localhost:8080/uppercase ; done
{"v":"FOO","err":null}
{"v":"BAR","err":null}
{"v":"BAZ","err":null}
```

```
listen=:8001 caller=logging.go:28 method=uppercase input=foo output=FOO err=null took=5.168µs
listen=:8080 caller=logging.go:28 method=uppercase input=foo output=FOO err=null took=4.39012ms
listen=:8002 caller=logging.go:28 method=uppercase input=bar output=BAR err=null took=5.445µs
listen=:8080 caller=logging.go:28 method=uppercase input=bar output=BAR err=null took=2.04831ms
listen=:8003 caller=logging.go:28 method=uppercase input=baz output=BAZ err=null took=3.285µs
listen=:8080 caller=logging.go:28 method=uppercase input=baz output=BAZ err=null took=1.388155ms
```

## Advanced topics

### Threading a context

The context object is used to carry information across conceptual boundaries in the scope of a single request.
In our example, we haven't yet threaded the context through our business logic.
But that's almost always a good idea.
It allows you to pass request-scoped information between business logic and middlewares,
 and is necessary for more sophisticated tasks like granular distributed tracing annotations.

Concretely, this means your business logic interfaces will look like

```go
type MyService interface {
	Foo(context.Context, string, int) (string, error)
	Bar(context.Context, string) error
	Baz(context.Context) (int, error)
}
```

### Request tracing

Once your infrastructure grows beyond a certain size, it becomes important to trace requests through multiple services, so you can identify and troubleshoot hotspots.
See [package tracing](https://github.com/go-kit/kit/blob/master/tracing) for more information.

### Creating a client package

It's possible to use Go kit to create a client package to your service, to make consuming your service easier from other Go programs.
Effectively, your client package will provide an implementation of your service interface, which invokes a remote service instance using a specific transport.
See [package addsvc/client](https://github.com/go-kit/kit/tree/master/examples/addsvc/client)
 or [package profilesvc/client](https://github.com/go-kit/kit/tree/master/examples/profilesvc/client)
 for examples.

## Other examples

### addsvc

[addsvc](https://github.com/go-kit/kit/blob/master/examples/addsvc) is the original example service.
It exposes a set of operations over **all supported transports**.
It's fully logged, instrumented, and uses Zipkin request tracing.
It also demonstrates how to create and use client packages.
It's a good example of a fully-featured Go kit service.

### profilesvc

[profilesvc](https://github.com/go-kit/kit/blob/master/examples/profilesvc)
 demonstrates how to use Go kit to build a REST-ish microservice.

### apigateway
[apigateway](https://github.com/go-kit/kit/blob/master/examples/apigateway/main.go)
 demonstrates how to implement the API gateway pattern,
 backed by a Consul service discovery system.

### shipping

[shipping](https://github.com/go-kit/kit/tree/master/examples/shipping)
 is a complete, "real-world" application composed of multiple microservices,
 based on Domain Driven Design principles.
