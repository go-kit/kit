// Package provider provides a factory-like abstraction for metrics backends.
// This package is provided specifically for the needs of the NY Times framework
// Gizmo. Most normal Go kit users shouldn't need to use it.
//
// Normally, if your microservice needs to support different metrics backends,
// you can simply do different construction based on a flag. For example,
//
//    var latency metrics.Histogram
//    var requests metrics.Counter
//    switch *metricsBackend {
//    case "prometheus":
//        latency = prometheus.NewSummaryVec(...)
//        requests = prometheus.NewCounterVec(...)
//    case "statsd":
//        s := statsd.New(...)
//        t := time.NewTicker(5*time.Second)
//        go s.SendLoop(t.C, "tcp", "statsd.local:8125")
//        requests = s.NewCounter(...)
//    default:
//        log.Fatal("unsupported metrics backend %q", *metricsBackend)
//    }
//
package provider
