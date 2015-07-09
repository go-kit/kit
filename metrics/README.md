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

A gauge for the number of goroutines currently running, exported via statsd.
```go
import (
	"net"
	"os"
	"runtime"
	"time"

	"github.com/go-kit/kit/metrics/statsd"
)

func main() {
	statsdWriter, err := net.Dial("udp", "127.0.0.1:8126")
	if err != nil {
		os.Exit(1)
	}

	reportingDuration := 5 * time.Second
	goroutines := statsd.NewGauge(statsdWriter, "total_goroutines", reportingDuration)
	for range time.Tick(reportingDuration) {
		goroutines.Set(float64(runtime.NumGoroutine()))
	}
}

```
