module github.com/go-kit/kit/examples

go 1.16

require (
	github.com/apache/thrift v0.14.1
	github.com/go-kit/kit v0.10.0
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/mux v1.7.3
	github.com/hashicorp/consul/api v1.3.0
	github.com/hashicorp/go-version v1.3.0 // indirect
	github.com/lightstep/lightstep-tracer-go v0.22.0
	github.com/nats-io/nats.go v1.11.0
	github.com/oklog/oklog v0.3.2
	github.com/oklog/run v1.1.0 // indirect
	github.com/opentracing/basictracer-go v1.1.0 // indirect
	github.com/opentracing/opentracing-go v1.2.0
	github.com/openzipkin-contrib/zipkin-go-opentracing v0.4.5
	github.com/openzipkin/zipkin-go v0.2.2
	github.com/pact-foundation/pact-go v1.0.4
	github.com/pborman/uuid v1.2.0
	github.com/prometheus/client_golang v1.5.1
	github.com/sony/gobreaker v0.4.1
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1
	google.golang.org/grpc v1.37.0
	sourcegraph.com/sourcegraph/appdash v0.0.0-20190731080439-ebfcffb1b5c0
)

replace github.com/go-kit/kit => ../
