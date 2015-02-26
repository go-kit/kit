# gokit

**gokit** is a working name for a **distributed programming toolkit** to serve the needs of the modern service-oriented enterprise.

- [GitHub repository](https://github.com/peterbourgon/gokit) -- RFCs, issues, PRs, etc.
- [go-kit mailing list](https://groups.google.com/forum/#!forum/go-kit)
- [Freenode](https://freenode.net) #gokit

## Motivation

See [the motivating blog post](http://peter.bourgon.org/go-kit) and, eventually, the video of the talk.

## Goals

- Operate in a heterogeneous SOA -- expect to interact with mostly non-gokit services
- RPC as the messaging pattern
- Pluggable serialization and transport -- not just JSON over HTTP
- Zipkin-compatible request tracing
- _more TODO_

## Non-goals

- Having opinions on deployment, orchestration, process supervision
- Having opinions on configuration passing -- flags vs. env vars vs. files vs. ...
- _more TODO_

## Related projects

Projects with a ★ have had particular influence on gokit's design.

### Service frameworks

- [Kite](https://github.com/koding/kite), a micro-service framework
- [go-micro](https://github.com/asim/go-micro), a microservices client/server library ★
- [gocircuit](https://github.com/gocircuit/circuit), dynamic cloud orchestration
- [gotalk](https://github.com/rsms/gotalk), async peer communication protocol &amp; library

### Individual components

- [grpc/grpc-go](https://github.com/grpc/grpc-go), HTTP/2 based RPC
- [afex/hystrix-go](https://github.com/afex/hystrix-go), client-side latency and fault tolerance library
- [streadway/handy](https://github.com/streadway/handy), net/http handler filters
- [rubyist/circuitbreaker](https://github.com/rubyist/circuitbreaker), circuit breaker library
- [spacemonkeygo/monitor](https://github.com/spacemonkeygo/monitor), data collection, monitoring, instrumentation, and Zipkin client library
- [mattheath/phosphor](https://github.com/mattheath/phosphor), distributed system tracing
- [codahale/lunk](https://github.com/codahale/lunk), structured logging in the style of Google's Dapper or Twitter's Zipkin
- [sourcegraph/appdash](https://github.com/sourcegraph/appdash), application tracing system based on Google's Dapper
- [eapache/go-resiliency](https://github.com/eapache/go-resiliency), resiliency patterns
- [FogCreek/logging](https://github.com/FogCreek/logging), a tagged style of logging
- [Sirupsen/logrus](https://github.com/Sirupsen/logrus), structured, pluggable logging for Go ★
- [mailgun/vulcand](https://github.com/mailgun/vulcand), prorammatic load balancer backed by etcd
- [vitess/rpcplus](https://godoc.org/code.google.com/p/vitess/go/rpcplus), package rpc + context.Context
- [pivotal-golang/lager](https://github.com/pivotal-golang/lager), an opinionated logging library

### Web frameworks

- [Gorilla](http://www.gorillatoolkit.org)
- [Revel](https://revel.github.io/)
- [Gin](https://gin-gonic.github.io/gin/)
- [Martini](https://github.com/go-martini/martini)
- [Negroni](https://github.com/codegangsta/negroni)
- [Goji](https://github.com/zenazn/goji)
- [Beego](http://beego.me/)

## Additional reading

- [Dapper, a Large-Scale Distributed Systems Tracing Infrastructure](http://research.google.com/pubs/pub36356.html) -- Google
- [Your Server as a Function](http://monkey.org/~marius/funsrv.pdf) (PDF) -- Twitter
- [Architecting for the Cloud](http://fr.slideshare.net/stonse/architecting-for-the-cloud-using-netflixoss-codemash-workshop-29852233) -- Netflix
