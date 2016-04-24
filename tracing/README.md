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
