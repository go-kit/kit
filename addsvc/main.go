package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	stdlog "log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/streadway/handy/cors"
	"github.com/streadway/handy/encoding"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	thriftadd "github.com/peterbourgon/gokit/addsvc/_thrift/gen-go/add"
	"github.com/peterbourgon/gokit/addsvc/pb"
	kitlog "github.com/peterbourgon/gokit/log"
	"github.com/peterbourgon/gokit/metrics"
	"github.com/peterbourgon/gokit/metrics/expvar"
	"github.com/peterbourgon/gokit/metrics/statsd"
	"github.com/peterbourgon/gokit/server"
	"github.com/peterbourgon/gokit/tracing/zipkin"
	jsoncodec "github.com/peterbourgon/gokit/transport/codec/json"
	httptransport "github.com/peterbourgon/gokit/transport/http"
)

func main() {
	var (
		httpAddr         = flag.String("http.addr", ":8001", "Address for HTTP (JSON) server")
		grpcAddr         = flag.String("grpc.addr", ":8002", "Address for gRPC server")
		thriftAddr       = flag.String("thrift.addr", ":8003", "Address for Thrift server")
		thriftProtocol   = flag.String("thrift.protocol", "binary", "binary, compact, json, simplejson")
		thriftBufferSize = flag.Int("thrift.buffer.size", 0, "0 for unbuffered")
		thriftFramed     = flag.Bool("thrift.framed", false, "true to enable framing")
	)
	flag.Parse()

	// `package metrics` domain
	requests := metrics.NewMultiCounter(
		expvar.NewCounter("requests"),
		statsd.NewCounter(ioutil.Discard, "requests", time.Second),
	)
	duration := metrics.NewMultiHistogram(
		expvar.NewHistogram("duration_ns", 0, 100000000, 3),
		statsd.NewHistogram(ioutil.Discard, "duration_ns", time.Second),
	)

	// `package tracing` domain
	zipkinHost := "some-host"             // TODO
	zipkinCollector := loggingCollector{} // TODO
	zipkinAddName := "ADD"                // is that right?
	zipkinAddSpanFunc := zipkin.NewSpanFunc(zipkinHost, zipkinAddName)

	// `package log` domain
	var logger kitlog.Logger
	logger = kitlog.NewPrefixLogger(kitlog.StdlibWriter{})
	logger = kitlog.With(logger, "ts", kitlog.DefaultTimestampUTC)
	kitlog.DefaultLogger = logger // for other gokit components
	stdlog.SetOutput(os.Stderr)   //
	stdlog.SetFlags(0)            // flags are handled in our logger

	// Our business and operational domain
	var a Add
	a = pureAdd
	a = logging(logger, a)

	// `package server` domain
	var e server.Endpoint
	e = makeEndpoint(a)
	e = zipkin.AnnotateEndpoint(zipkinAddSpanFunc, zipkinCollector)(e)
	// e = someother.Middleware(arg1, arg2)(e)

	// Mechanical stuff
	root := context.Background()
	errc := make(chan error)

	go func() {
		errc <- interrupt()
	}()

	// Transport: gRPC
	go func() {
		ln, err := net.Listen("tcp", *grpcAddr)
		if err != nil {
			errc <- err
			return
		}
		s := grpc.NewServer() // uses its own context?
		field := metrics.Field{Key: "transport", Value: "grpc"}

		var addServer pb.AddServer
		addServer = grpcBinding{e}
		addServer = grpcInstrument(requests.With(field), duration.With(field))(addServer)

		pb.RegisterAddServer(s, addServer)
		logger.Log("msg", "gRPC server started", "addr", *grpcAddr)
		errc <- s.Serve(ln)
	}()

	// Transport: HTTP (JSON)
	go func() {
		ctx, cancel := context.WithCancel(root)
		defer cancel()

		field := metrics.Field{Key: "transport", Value: "http"}
		before := httptransport.Before(zipkin.ToContext(zipkin.FromHTTP(zipkinAddSpanFunc)))
		after := httptransport.After(httptransport.SetContentType("application/json"))

		var handler http.Handler
		handler = httptransport.NewBinding(ctx, reflect.TypeOf(request{}), jsoncodec.New(), e, before, after)
		handler = encoding.Gzip(handler)
		handler = cors.Middleware(cors.Config{})(handler)
		handler = httpInstrument(requests.With(field), duration.With(field))(handler)

		mux := http.NewServeMux()
		mux.Handle("/add", handler)
		logger.Log("msg", "HTTP server started", "addr", *httpAddr)
		errc <- http.ListenAndServe(*httpAddr, mux)
	}()

	// Transport: Thrift
	go func() {
		ctx, cancel := context.WithCancel(root)
		defer cancel()

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

		field := metrics.Field{Key: "transport", Value: "thrift"}

		var service thriftadd.AddService
		service = thriftBinding{ctx, e}
		service = thriftInstrument(requests.With(field), duration.With(field))(service)

		logger.Log("msg", "Thrift server started", "addr", *thriftAddr)
		errc <- thrift.NewTSimpleServer4(
			thriftadd.NewAddServiceProcessor(service),
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

type loggingCollector struct{}

func (loggingCollector) Collect(s *zipkin.Span) error {
	kitlog.With(kitlog.DefaultLogger, "caller", kitlog.DefaultCaller).Log(
		"trace_id", s.TraceID(),
		"span_id", s.SpanID(),
		"parent_span_id", s.ParentSpanID(),
	)
	return nil
}
