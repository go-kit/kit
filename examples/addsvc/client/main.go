package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	grpcclient "github.com/go-kit/kit/examples/addsvc/client/grpc"
	httpjsonclient "github.com/go-kit/kit/examples/addsvc/client/httpjson"
	netrpcclient "github.com/go-kit/kit/examples/addsvc/client/netrpc"
	thriftclient "github.com/go-kit/kit/examples/addsvc/client/thrift"
	"github.com/go-kit/kit/loadbalancer"
	"github.com/go-kit/kit/loadbalancer/static"
	"github.com/go-kit/kit/log"
	kitot "github.com/go-kit/kit/tracing/opentracing"
	"github.com/lightstep/lightstep-tracer-go"
	"github.com/opentracing/opentracing-go"
	appdashot "github.com/sourcegraph/appdash/opentracing"
	"sourcegraph.com/sourcegraph/appdash"
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

		// Two OpenTracing backends (to demonstrate how they can be interchanged):
		appdashHostport      = flag.String("appdash_hostport", "", "Enable Appdash tracing via an Appdash server host:port")
		lightstepAccessToken = flag.String("lightstep_access_token", "", "Enable LightStep tracing via a LightStep access token")
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

	// Set up OpenTracing
	var tracer opentracing.Tracer
	{
		if len(*appdashHostport) > 0 {
			tracer = appdashot.NewTracer(appdash.NewRemoteCollector(*appdashHostport))
		}
		if len(*lightstepAccessToken) > 0 {
			if tracer != nil {
				panic("Attempted to configure multiple OpenTracing implementations")
			}
			tracer = lightstep.NewTracer(lightstep.Options{
				AccessToken: *lightstepAccessToken,
			})
		}
		if tracer == nil {
			tracer = opentracing.GlobalTracer() // the noop tracer
		}
	}

	var (
		instances                 []string
		sumFactory, concatFactory loadbalancer.Factory
	)

	switch *transport {
	case "grpc":
		instances = strings.Split(*grpcAddrs, ",")
		sumFactory = grpcclient.NewSumEndpointFactory(tracer)
		concatFactory = grpcclient.NewConcatEndpointFactory(tracer)

	case "httpjson":
		instances = strings.Split(*httpAddrs, ",")
		for i, rawurl := range instances {
			if !strings.HasPrefix("http", rawurl) {
				instances[i] = "http://" + rawurl
			}
		}
		sumFactory = httpjsonclient.NewSumEndpointFactory(tracer)
		concatFactory = httpjsonclient.NewConcatEndpointFactory(tracer)

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

	if len(*lightstepAccessToken) > 0 {
		lightstep.FlushLightStepTracer(tracer)
	}
}

func buildEndpoint(tracer opentracing.Tracer, operationName string, instances []string, factory loadbalancer.Factory, seed int64, logger log.Logger) endpoint.Endpoint {
	publisher := static.NewPublisher(instances, factory, logger)
	random := loadbalancer.NewRandom(publisher, seed)
	endpoint := loadbalancer.Retry(10, 10*time.Second, random)
	return kitot.TraceClient(tracer, operationName)(endpoint)
}
