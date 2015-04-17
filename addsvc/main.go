package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/streadway/handy/cors"
	"github.com/streadway/handy/encoding"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	thriftadd "github.com/peterbourgon/gokit/addsvc/_thrift/gen-go/add"
	"github.com/peterbourgon/gokit/addsvc/pb"
	"github.com/peterbourgon/gokit/metrics"
	"github.com/peterbourgon/gokit/metrics/expvar"
	"github.com/peterbourgon/gokit/metrics/statsd"
	"github.com/peterbourgon/gokit/server"
	"github.com/peterbourgon/gokit/tracing/zipkin"
	kithttp "github.com/peterbourgon/gokit/transport/http"
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

	// Our business and operational domain
	var a Add
	a = pureAdd
	a = logging(logWriter{}, a)

	// `package server` domain
	var e server.Endpoint
	e = makeEndpoint(a)
	// e = server.ChainableEnhancement(arg1, arg2, e)

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
	zipkinHost := "some-host"                // TODO
	zipkinCollector := zipkin.NopCollector{} // TODO
	zipkinSpanFunc := zipkin.NewSpanFunc(zipkinHost, zipkinCollector)

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
		log.Printf("gRPC server on %s", *grpcAddr)
		errc <- s.Serve(ln)
	}()

	// Transport: HTTP (JSON)
	go func() {
		ctx, cancel := context.WithCancel(root)
		defer cancel()

		field := metrics.Field{Key: "transport", Value: "http"}
		before := kithttp.Before(zipkin.ToContext(zipkin.FromHTTP(zipkinSpanFunc)))
		after := kithttp.After(kithttp.SetContentType("application/json"))

		var handler http.Handler
		handler = kithttp.NewBinding(ctx, jsonCodec{}, e, before, after)
		handler = encoding.Gzip(handler)
		handler = cors.Middleware(cors.Config{})(handler)
		handler = httpInstrument(requests.With(field), duration.With(field))(handler)

		mux := http.NewServeMux()
		mux.Handle("/add", handler)
		log.Printf("HTTP server on %s", *httpAddr)
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

		log.Printf("Thrift server on %s", *thriftAddr)
		errc <- thrift.NewTSimpleServer4(
			thriftadd.NewAddServiceProcessor(service),
			transport,
			transportFactory,
			protocolFactory,
		).Serve()
	}()

	log.Fatal(<-errc)
}

type logWriter struct{}

func (logWriter) Write(p []byte) (int, error) {
	log.Printf("%s", p)
	return len(p), nil
}

func interrupt() error {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	return fmt.Errorf("%s", <-c)
}
