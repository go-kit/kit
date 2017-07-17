package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/oklog/oklog/pkg/group"
	stdopentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"

	addpb "github.com/go-kit/kit/examples/addsvc2/pb"
	addendpoint "github.com/go-kit/kit/examples/addsvc2/pkg/endpoint"
	addservice "github.com/go-kit/kit/examples/addsvc2/pkg/service"
	addtransport "github.com/go-kit/kit/examples/addsvc2/pkg/transport"
	addthrift "github.com/go-kit/kit/examples/addsvc2/thrift/gen-go/addsvc"
)

func main() {
	var (
		debugAddr        = flag.String("debug.addr", ":8080", "Debug and metrics listen address")
		httpAddr         = flag.String("http-addr", ":8081", "HTTP listen address")
		grpcAddr         = flag.String("grpc-addr", ":8082", "gRPC listen address")
		thriftAddr       = flag.String("thrift-addr", ":8082", "Thrift listen address")
		thriftProtocol   = flag.String("thrift.protocol", "binary", "binary, compact, json, simplejson")
		thriftBufferSize = flag.Int("thrift.buffer.size", 0, "0 for unbuffered")
		thriftFramed     = flag.Bool("thrift.framed", false, "true to enable framing")
		zipkinURL        = flag.String("zipkin-url", "", "Zipkin collector URL e.g. http://localhost:9411/api/v1/spans")
	)
	flag.Parse()

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	var tracer stdopentracing.Tracer
	{
		if *zipkinURL != "" {
			logger.Log("zipkin", *zipkinURL)
			collector, err := zipkin.NewHTTPCollector(*zipkinURL)
			if err != nil {
				logger.Log("err", err)
				os.Exit(1)
			}
			defer collector.Close()
			var (
				debug       = false
				hostPort    = "localhost:80"
				serviceName = "addsvc"
			)
			tracer, err = zipkin.NewTracer(zipkin.NewRecorder(
				collector, debug, hostPort, serviceName,
			))
			if err != nil {
				logger.Log("err", err)
				os.Exit(1)
			}
		} else {
			tracer = stdopentracing.GlobalTracer() // no-op
		}
	}

	// Our metrics are dependencies, here we create them.
	var ints, chars metrics.Counter
	{
		// Business-level metrics.
		ints = prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "example",
			Subsystem: "addsvc",
			Name:      "integers_summed",
			Help:      "Total count of integers summed via the Sum method.",
		}, []string{})
		chars = prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "example",
			Subsystem: "addsvc",
			Name:      "characters_concatenated",
			Help:      "Total count of characters concatenated via the Concat method.",
		}, []string{})
	}
	var duration metrics.Histogram
	{
		// Endpoint-level metrics.
		duration = prometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "example",
			Subsystem: "addsvc",
			Name:      "request_duration_seconds",
			Help:      "Request duration in seconds.",
		}, []string{"method", "success"})
	}

	var (
		service       = addservice.New(logger, ints, chars)
		endpoints     = addendpoint.New(service, logger, duration, tracer)
		httpHandler   = addtransport.NewHTTPHandler(context.Background(), endpoints, logger, tracer)
		grpcServer    = addtransport.MakeGRPCServer(endpoints, tracer, logger)
		thriftHandler = addtransport.MakeThriftHandler(context.Background(), endpoints)
	)

	var g group.Group
	{
		debugListener, err := net.Listen("tcp", *debugAddr)
		if err != nil {
			logger.Log("transport", "debug/HTTP", "during", "Listen", "err", err)
			os.Exit(1)
		}
		g.Add(func() error {
			logger.Log("transport", "debug/HTTP", "addr", *debugAddr)
			return http.Serve(debugListener, http.DefaultServeMux)
		}, func(error) {
			debugListener.Close()
		})
	}
	{
		httpListener, err := net.Listen("tcp", *httpAddr)
		if err != nil {
			logger.Log("transport", "HTTP", "during", "Listen", "err", err)
			os.Exit(1)
		}
		g.Add(func() error {
			logger.Log("transport", "HTTP", "addr", *httpAddr)
			return http.Serve(httpListener, httpHandler)
		}, func(error) {
			httpListener.Close()
		})
	}
	{
		grpcListener, err := net.Listen("tcp", *grpcAddr)
		if err != nil {
			logger.Log("transport", "gRPC", "during", "Listen", "err", err)
			os.Exit(1)
		}
		g.Add(func() error {
			logger.Log("transport", "gRPC", "addr", *grpcAddr)
			baseServer := grpc.NewServer()
			addpb.RegisterAddServer(baseServer, grpcServer)
			return baseServer.Serve(grpcListener)
		}, func(error) {
			grpcListener.Close()
		})
	}
	{
		thriftSocket, err := thrift.NewTServerSocket(*thriftAddr)
		if err != nil {
			logger.Log("transport", "Thrift", "during", "Listen", "err", err)
			os.Exit(1)
		}
		g.Add(func() error {
			logger.Log("transport", "Thrift", "addr", *thriftAddr)
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
				return fmt.Errorf("invalid Thrift protocol %q", *thriftProtocol)
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
			return thrift.NewTSimpleServer4(
				addthrift.NewAddServiceProcessor(thriftHandler),
				thriftSocket,
				transportFactory,
				protocolFactory,
			).Serve()
		}, func(error) {
			thriftSocket.Close()
		})
	}
	{
		cancelInterrupt := make(chan struct{})
		g.Add(func() error {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			select {
			case sig := <-c:
				return fmt.Errorf("received signal %s", sig)
			case <-cancelInterrupt:
				return nil
			}
		}, func(error) {
			close(cancelInterrupt)
		})
	}
	logger.Log("exit", g.Run())
}
