# package metrics

`package metrics` provides a set of uniform interfaces for service instrumentation.
It has **[counters][]**, **[gauges][]**, and **[histograms][]**,
 and provides adapters to popular metrics packages, like **[expvar][]**, **[statsd][]**, and **[Prometheus][]**.

[counters]: http://prometheus.io/docs/concepts/metric_types/#counter
[gauges]: http://prometheus.io/docs/concepts/metric_types/#gauge
[histograms]: http://prometheus.io/docs/concepts/metric_types/#histogram
[expvar]: https://golang.org/pkg/expvar
[statsd]: https://github.com/etsy/statsd
[Prometheus]: http://prometheus.io

## Rationale

TODO

## Usage

A simple counter, exported via expvar.

```go
import "github.com/go-kit/kit/metrics/expvar"

func main() {
	myCount := expvar.NewCounter("my_count")
	myCount.Add(1)
}
```

A histogram for request duration, exported via a Prometheus summary with
dynamically-computed quantiles.

```go
import (
	stdprometheus "github.com/prometheus/client_golang/prometheus"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	"github.com/go-kit/kit/metrics/statsd"
)

var requestDuration = prometheus.NewSummary(stdprometheus.SummaryOpts{
	Namespace: "myservice",
	Subsystem: "api",
	Name:      "request_duration_nanoseconds_count",
	Help:      "Total time spent serving requests.",
}, []string{})

func handleRequest() {
	defer func(begin time.Time) { requestDuration.Observe(time.Since(begin)) }(time.Now())
	// handle request
}
```
