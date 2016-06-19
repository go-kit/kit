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

Code instrumentation is absolutely essential to achieve [observability][] into a distributed system.
Metrics and instrumentation tools have coalesced around a few well-defined idioms.
`package metrics` provides a common, minimal interface those idioms for service authors.

Using this interface allows library authors to easily support exporting a wide
variety of metrics without committing to any single metrics provider.

[observability]: https://speakerdeck.com/mattheath/observability-in-micro-service-architectures

## Usage

A simple counter, exported via expvar.

```go
import (
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/expvar"
)

func main() {
	var myCount metrics.Counter
	myCount = expvar.NewCounter("my_count")
	myCount.Add(1)
}

```

A histogram for request duration, exported via a Prometheus summary with
dynamically-computed quantiles.

```go
import (
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

var requestDuration = prometheus.NewSummary(stdprometheus.SummaryOpts{
	Namespace: "myservice",
	Subsystem: "api",
	Name:      "request_duration_nanoseconds_count",
	Help:      "Total time spent serving requests.",
}, []string{})

func handleRequest(requestDur metrics.Histogram) {
	defer func(begin time.Time) { requestDur.Observe(int64(time.Since(begin))) }(time.Now())
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
		panic(err)
	}

	reportInterval := 5 * time.Second
    var goroutines metrics.Gauge
	goroutines = statsd.NewGauge(statsdWriter, "total_goroutines", reportInterval)
    exportGoroutines(goroutines, reportInterval)
}

func exportGoroutines(g metrics.Gauge, interval time.Duration) {
    for range time.Tick(reportInterval) {
        goroutines.Set(float64(runtime.NumGoroutine()))
    }
}
```
