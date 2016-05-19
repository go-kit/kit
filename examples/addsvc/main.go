package main

import (
	"flag"
	"fmt"
	stdlog "log"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
	zipkin "github.com/basvanbeek/zipkin-go-opentracing"
	"github.com/lightstep/lightstep-tracer-go"
	"github.com/opentracing/opentracing-go"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	appdashot "github.com/sourcegraph/appdash/opentracing"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"sourcegraph.com/sourcegraph/appdash"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/examples/addsvc/pb"
	"github.com/go-kit/kit/examples/addsvc/server"
	thriftadd "github.com/go-kit/kit/examples/addsvc/thrift/gen-go/add"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/expvar"
	"github.com/go-kit/kit/metrics/prometheus"
	kitot "github.com/go-kit/kit/tracing/opentracing"
	httptransport "github.com/go-kit/kit/transport/http"
)

func main() {
	// Flag domain. Note that gRPC transitively registers flags via its import
	// of glog. So, we define a new flag set, to keep those domains distinct.
	fs := flag.NewFlagSet("", flag.ExitOnError)
	var (
		debugAddr        = fs.String("debug.addr", ":8000", "Address for HTTP debug/instrumentation server")
		httpAddr         = fs.String("http.addr", ":8001", "Address for HTTP (JSON) server")
		grpcAddr         = fs.String("grpc.addr", ":8002", "Address for gRPC server")
		netrpcAddr       = fs.String("netrpc.addr", ":8003", "Address for net/rpc server")
		thriftAddr       = fs.String("thrift.addr", ":8004", "Address for Thrift server")
		thriftProtocol   = fs.String("thrift.protocol", "binary", "binary, compact, json, simplejson")
		thriftBufferSize = fs.Int("thrift.buffer.size", 0, "0 for unbuffered")
		thriftFramed     = fs.Bool("thrift.framed", false, "true to enable framing")

		// Supported OpenTracing backends
		zipkinAddr           = fs.String("zipkin.kafka.addr", "", "Enable Zipkin tracing via a Kafka server host:port")
		appdashAddr          = fs.String("appdash.addr", "", "Enable Appdash tracing via an Appdash server host:port")
		lightstepAccessToken = fs.String("lightstep.token", "", "Enable LightStep tracing via a LightStep access token")
	)
	flag.Usage = fs.Usage // only show our flags
	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}

	// package log
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.NewContext(logger).With("ts", log.DefaultTimestampUTC).With("caller", log.DefaultCaller)
		stdlog.SetFlags(0)                             // flags are handled by Go kit's logger
		stdlog.SetOutput(log.NewStdlibAdapter(logger)) // redirect anything using stdlib log to us
	}

	// package metrics
	var requestDuration metrics.TimeHistogram
	{
		requestDuration = metrics.NewTimeHistogram(time.Nanosecond, metrics.NewMultiHistogram(
			"request_duration_ns",
			expvar.NewHistogram("request_duration_ns", 0, 5e9, 1, 50, 95, 99),
			prometheus.NewSummary(stdprometheus.SummaryOpts{
				Namespace: "myorg",
				Subsystem: "addsvc",
				Name:      "duration_ns",
				Help:      "Request duration in nanoseconds.",
			}, []string{"method"}),
		))
	}

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
				zipkin.KafkaLogger(logger),
			)
			if err != nil {
				logger.Log("err", "unable to create kafka collector")
				os.Exit(1)
			}
			tracer, err = zipkin.NewTracer(
				zipkin.NewRecorder(collector, false, "localhost:80", "addsvc"),
			)
			if err != nil {
				logger.Log("err", "unable to create zipkin tracer")
				os.Exit(1)
			}
		case *appdashAddr == "" && *lightstepAccessToken == "" && *zipkinAddr == "":
			tracer = opentracing.GlobalTracer() // no-op
		default:
			panic("specify a single -appdash.addr, -lightstep.access.token or -zipkin.kafka.addr")
		}
	}

	// Business domain
	var svc server.AddService
	{
		svc = pureAddService{}
		svc = loggingMiddleware{svc, logger}
		svc = instrumentingMiddleware{svc, requestDuration}
	}

	// Mechanical stuff
	rand.Seed(time.Now().UnixNano())
	root := context.Background()
	errc := make(chan error)

	go func() {
		errc <- interrupt()
	}()

	// Debug/instrumentation
	go func() {
		transportLogger := log.NewContext(logger).With("transport", "debug")
		transportLogger.Log("addr", *debugAddr)
		errc <- http.ListenAndServe(*debugAddr, nil) // DefaultServeMux
	}()

	// Transport: HTTP/JSON
	go func() {
		var (
			transportLogger = log.NewContext(logger).With("transport", "HTTP/JSON")
			tracingLogger   = log.NewContext(transportLogger).With("component", "tracing")
			mux             = http.NewServeMux()
			sum, concat     endpoint.Endpoint
		)

		sum = makeSumEndpoint(svc)
		sum = kitot.TraceServer(tracer, "sum")(sum)
		mux.Handle("/sum", httptransport.NewServer(
			root,
			sum,
			server.DecodeSumRequest,
			server.EncodeSumResponse,
			httptransport.ServerErrorLogger(transportLogger),
			httptransport.ServerBefore(kitot.FromHTTPRequest(tracer, "sum", tracingLogger)),
		))

		concat = makeConcatEndpoint(svc)
		concat = kitot.TraceServer(tracer, "concat")(concat)
		mux.Handle("/concat", httptransport.NewServer(
			root,
			concat,
			server.DecodeConcatRequest,
			server.EncodeConcatResponse,
			httptransport.ServerErrorLogger(transportLogger),
			httptransport.ServerBefore(kitot.FromHTTPRequest(tracer, "concat", tracingLogger)),
		))

		transportLogger.Log("addr", *httpAddr)
		errc <- http.ListenAndServe(*httpAddr, mux)
	}()

	// Transport: gRPC
	go func() {
		transportLogger := log.NewContext(logger).With("transport", "gRPC")
		tracingLogger := log.NewContext(transportLogger).With("component", "tracing")
		ln, err := net.Listen("tcp", *grpcAddr)
		if err != nil {
			errc <- err
			return
		}
		s := grpc.NewServer() // uses its own, internal context
		pb.RegisterAddServer(s, newGRPCBinding(root, tracer, svc, tracingLogger))
		transportLogger.Log("addr", *grpcAddr)
		errc <- s.Serve(ln)
	}()

	// Transport: net/rpc
	go func() {
		transportLogger := log.NewContext(logger).With("transport", "net/rpc")
		s := rpc.NewServer()
		if err := s.RegisterName("addsvc", netrpcBinding{svc}); err != nil {
			errc <- err
			return
		}
		s.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)
		transportLogger.Log("addr", *netrpcAddr)
		errc <- http.ListenAndServe(*netrpcAddr, s)
	}()

	// Transport: Thrift
	go func() {
		var protocolFactory thrift.TProtocolFactory
		switch *thriftProtocol {
		case "binary":
			protocolFactory = thrift.NewTBinaryProtocolFactoryDefault()
		case "compact":
			protocolFactory = thrift.NewTCompactProtocolFactory()
		case "json":
			protocolFactory = thrift.NewTJSONProtocolFactory()
		case "simplejson":
			protocolFactory = thrift.NewTSimpleJSONProtocolFactory()
		default:
			errc <- fmt.Errorf("invalid Thrift protocol %q", *thriftProtocol)
			return
		}
		var transportFactory thrift.TTransportFactory
		if *thriftBufferSize > 0 {
			transportFactory = thrift.NewTBufferedTransportFactory(*thriftBufferSize)
		} else {
			transportFactory = thrift.NewTTransportFactory()
		}
		if *thriftFramed {
			transportFactory = thrift.NewTFramedTransportFactory(transportFactory)
		}
		transport, err := thrift.NewTServerSocket(*thriftAddr)
		if err != nil {
			errc <- err
			return
		}
		transportLogger := log.NewContext(logger).With("transport", "thrift")
		transportLogger.Log("addr", *thriftAddr)
		errc <- thrift.NewTSimpleServer4(
			thriftadd.NewAddServiceProcessor(thriftBinding{svc}),
			transport,
			transportFactory,
			protocolFactory,
		).Serve()
	}()

	logger.Log("fatal", <-errc)
}

func interrupt() error {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	return fmt.Errorf("%s", <-c)
}
