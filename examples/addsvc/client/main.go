package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	zipkin "github.com/basvanbeek/zipkin-go-opentracing"
	"github.com/lightstep/lightstep-tracer-go"
	"github.com/opentracing/opentracing-go"
	appdashot "github.com/sourcegraph/appdash/opentracing"
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/appdash"

	"github.com/go-kit/kit/endpoint"
	grpcclient "github.com/go-kit/kit/examples/addsvc/client/grpc"
	httpjsonclient "github.com/go-kit/kit/examples/addsvc/client/httpjson"
	netrpcclient "github.com/go-kit/kit/examples/addsvc/client/netrpc"
	thriftclient "github.com/go-kit/kit/examples/addsvc/client/thrift"
	"github.com/go-kit/kit/loadbalancer"
	"github.com/go-kit/kit/loadbalancer/static"
	"github.com/go-kit/kit/log"
	kitot "github.com/go-kit/kit/tracing/opentracing"
)

func main() {
	var (
		transport        = flag.String("transport", "httpjson", "httpjson, grpc, netrpc, thrift")
		httpAddrs        = flag.String("http.addrs", "localhost:8001", "Comma-separated list of addresses for HTTP (JSON) servers")
		grpcAddrs        = flag.String("grpc.addrs", "localhost:8002", "Comma-separated list of addresses for gRPC servers")
		netrpcAddrs      = flag.String("netrpc.addrs", "localhost:8003", "Comma-separated list of addresses for net/rpc servers")
		thriftAddrs      = flag.String("thrift.addrs", "localhost:8004", "Comma-separated list of addresses for Thrift servers")
		thriftProtocol   = flag.String("thrift.protocol", "binary", "binary, compact, json, simplejson")
		thriftBufferSize = flag.Int("thrift.buffer.size", 0, "0 for unbuffered")
		thriftFramed     = flag.Bool("thrift.framed", false, "true to enable framing")

		// Three OpenTracing backends (to demonstrate how they can be interchanged):
		zipkinAddr           = flag.String("zipkin.kafka.addr", "", "Enable Zipkin tracing via a Kafka Collector host:port")
		appdashAddr          = flag.String("appdash.addr", "", "Enable Appdash tracing via an Appdash server host:port")
		lightstepAccessToken = flag.String("lightstep.token", "", "Enable LightStep tracing via a LightStep access token")
	)
	flag.Parse()
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "\n%s [flags] method arg1 arg2\n\n", filepath.Base(os.Args[0]))
		flag.Usage()
		os.Exit(1)
	}

	randomSeed := time.Now().UnixNano()

	root := context.Background()
	method, s1, s2 := flag.Arg(0), flag.Arg(1), flag.Arg(2)

	var logger log.Logger
	logger = log.NewLogfmtLogger(os.Stdout)
	logger = log.NewContext(logger).With("caller", log.DefaultCaller)
	logger = log.NewContext(logger).With("transport", *transport)
	tracingLogger := log.NewContext(logger).With("component", "tracing")

	// Set up OpenTracing
	var tracer opentracing.Tracer
	{
		switch {
		case *appdashAddr != "" && *lightstepAccessToken == "" && *zipkinAddr == "":
			tracer = appdashot.NewTracer(appdash.NewRemoteCollector(*appdashAddr))
		case *appdashAddr == "" && *lightstepAccessToken != "" && *zipkinAddr == "":
			tracer = lightstep.NewTracer(lightstep.Options{
				AccessToken: *lightstepAccessToken,
			})
			defer lightstep.FlushLightStepTracer(tracer)
		case *appdashAddr == "" && *lightstepAccessToken == "" && *zipkinAddr != "":
			collector, err := zipkin.NewKafkaCollector(
				strings.Split(*zipkinAddr, ","),
				zipkin.KafkaLogger(tracingLogger),
			)
			if err != nil {
				tracingLogger.Log("err", "unable to create kafka collector")
				os.Exit(1)
			}
			tracer, err = zipkin.NewTracer(
				zipkin.NewRecorder(collector, false, "localhost:8000", "addsvc-client"),
			)
			if err != nil {
				tracingLogger.Log("err", "unable to create zipkin tracer")
				os.Exit(1)
			}
		case *appdashAddr == "" && *lightstepAccessToken == "" && *zipkinAddr == "":
			tracer = opentracing.GlobalTracer() // no-op
		default:
			panic("specify a single -appdash.addr, -lightstep.access.token or -zipkin.kafka.addr")
		}
	}

	var (
		instances                 []string
		sumFactory, concatFactory loadbalancer.Factory
	)

	switch *transport {
	case "grpc":
		instances = strings.Split(*grpcAddrs, ",")
		sumFactory = grpcclient.MakeSumEndpointFactory(tracer, tracingLogger)
		concatFactory = grpcclient.MakeConcatEndpointFactory(tracer, tracingLogger)

	case "httpjson":
		instances = strings.Split(*httpAddrs, ",")
		for i, rawurl := range instances {
			if !strings.HasPrefix("http", rawurl) {
				instances[i] = "http://" + rawurl
			}
		}
		sumFactory = httpjsonclient.MakeSumEndpointFactory(tracer, tracingLogger)
		concatFactory = httpjsonclient.MakeConcatEndpointFactory(tracer, tracingLogger)

	case "netrpc":
		instances = strings.Split(*netrpcAddrs, ",")
		sumFactory = netrpcclient.SumEndpointFactory
		concatFactory = netrpcclient.ConcatEndpointFactory

	case "thrift":
		instances = strings.Split(*thriftAddrs, ",")
		thriftClient := thriftclient.New(*thriftProtocol, *thriftBufferSize, *thriftFramed, logger)
		sumFactory = thriftClient.SumEndpoint
		concatFactory = thriftClient.ConcatEndpoint

	default:
		logger.Log("err", "invalid transport")
		os.Exit(1)
	}

	sum := buildEndpoint(tracer, "sum", instances, sumFactory, randomSeed, logger)
	concat := buildEndpoint(tracer, "concat", instances, concatFactory, randomSeed, logger)

	svc := newClient(root, sum, concat, logger)

	begin := time.Now()
	switch method {
	case "sum":
		a, _ := strconv.Atoi(s1)
		b, _ := strconv.Atoi(s2)
		v := svc.Sum(a, b)
		logger.Log("method", "sum", "a", a, "b", b, "v", v, "took", time.Since(begin))

	case "concat":
		a, b := s1, s2
		v := svc.Concat(a, b)
		logger.Log("method", "concat", "a", a, "b", b, "v", v, "took", time.Since(begin))

	default:
		logger.Log("err", "invalid method "+method)
		os.Exit(1)
	}
	// wait for collector
	time.Sleep(2 * time.Second)
}

func buildEndpoint(tracer opentracing.Tracer, operationName string, instances []string, factory loadbalancer.Factory, seed int64, logger log.Logger) endpoint.Endpoint {
	publisher := static.NewPublisher(instances, factory, logger)
	random := loadbalancer.NewRandom(publisher, seed)
	endpoint := loadbalancer.Retry(10, 10*time.Second, random)
	return kitot.TraceClient(tracer, operationName)(endpoint)
}
