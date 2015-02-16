---
RFC: 003
Author: Peter Bourgon <peter@bourgon.org>
Status: Draft
---

# package metrics

## Motivation

http://peter.bourgon.org/go-kit/#package-metrics

## Scope

- Package metrics SHALL implement Gauges, Counters, and Histograms.

- Each metric type SHALL allow observations with an unlimited number of key/value field pairs,
  similar to [package log](https://github.com/peterbourgon/gokit/blob/master/rfc/rfc004-package-log.md).

- Counter SHALL be an increment-only counter of type uint64.

- Gauge SHALL be an arbitrarily-settable register of type int64.

- Histogram SHALL collect observations of type int64.

- These interfaces SHALL be the primary and exclusive API for metrics.

- We SHALL provide a variety of implementations of each interface that act as a
  bridge to different backends: expvar, Graphite, statsd, Prometheus, etc.

- Each metric backend MAY provide additional value-add behaviors. For example,
  a backend for Histogram may bucket observations according to quantile and
  calculate additional, derived statistics.


## Implementation

https://github.com/peterbourgon/gokit/tree/master/metrics

### Gauge

```go
type Gauge interface {
	With(Field) Gauge
	Set(value int64)
	Add(delta int64)
}
```

### Counter

```go
type Counter interface {
	With(Field) Counter
	Add(delta uint64)
}
```

### Histogram

```go
type Histogram interface {
	With(Field) Histogram
	Observe(int64)
}
```
