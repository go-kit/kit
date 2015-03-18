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

	"git.apache.org/thrift.git/lib/go/thrift"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/peterbourgon/gokit/addsvc/pb"
	thriftadd "github.com/peterbourgon/gokit/addsvc/thrift/gen-go/add"
	"github.com/peterbourgon/gokit/metrics"
	"github.com/peterbourgon/gokit/metrics/expvar"
	"github.com/peterbourgon/gokit/metrics/statsd"
	"github.com/peterbourgon/gokit/server"
	"github.com/peterbourgon/gokit/server/zipkin"
	"github.com/peterbourgon/gokit/transport/http/cors"
)

func main() {
	var (
		httpJSONAddr     = flag.String("http.json.addr", ":8001", "Address for HTTP/JSON server")
		grpcTCPAddr      = flag.String("grpc.tcp.addr", ":8002", "Address for gRPC (TCP) server")
		thriftTCPAddr    = flag.String("thrift.tcp.addr", ":8003", "Address for Thrift (TCP) server")
		thriftProtocol   = flag.String("thrift.protocol", "binary", "binary, compact, json, simplejson")
		thriftBufferSize = flag.Int("thrift.transport.bufsz", 0, "Thrift transport buffer size (0 = unbuffered)")
		thriftFramed     = flag.Bool("thrift.framed", false, "use framed transport")
	)
	flag.Parse()

	// Our business and operational domain
	var a Add
	a = pureAdd
	a = logging(logWriter{}, a)

	// `package server` domain
	var e server.Endpoint
	e = makeEndpoint(a)
	e = server.Gate(zipkin.RequireInContext)(e) // must have Zipkin headers
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

	// Mechanical stuff
	root := context.Background()
	errc := make(chan error)

	go func() {
		errc <- interrupt()
	}()

	// Transport: gRPC
	go func() {
		ln, err := net.Listen("tcp", *grpcTCPAddr)
		if err != nil {
			errc <- err
			return
		}
		s := grpc.NewServer() // uses its own context?
		field := metrics.Field{Key: "transport", Value: "grpc"}

		var addServer pb.AddServer
		addServer = grpcBinding{e}
		addServer = grpcInstrument(requests.With(field), duration.With(field))(addServer)
		// Note that this will always fail, because the Endpoint is gated on
		// Zipkin headers, and we don't extract them from the gRPC request.

		pb.RegisterAddServer(s, addServer)
		log.Printf("gRPC server on TCP %s", *grpcTCPAddr)
		errc <- s.Serve(ln)
	}()

	// Transport: HTTP/JSON
	go func() {
		ctx, cancel := context.WithCancel(root)
		defer cancel()
		mux := http.NewServeMux()
		field := metrics.Field{Key: "transport", Value: "http"}

		var handler http.Handler
		handler = httpBinding{ctx, jsonCodec{}, "application/json", e}
		handler = httpInstrument(requests.With(field), duration.With(field))(handler)
		handler = cors.Middleware(cors.MaxAge(5 * time.Minute))(handler)

		mux.Handle("/add", handler)
		log.Printf("HTTP/JSON server on %s", *httpJSONAddr)
		errc <- http.ListenAndServe(*httpJSONAddr, mux)
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

		transport, err := thrift.NewTServerSocket(*thriftTCPAddr)
		if err != nil {
			errc <- err
			return
		}

		var handler thriftadd.AddService
		handler = thriftBinding{ctx, e}
		field := metrics.Field{Key: "transport", Value: "thrift"}
		handler = thriftInstrument(requests.With(field), duration.With(field))(handler)
		// Note that this will always fail, because the Endpoint is gated on
		// Zipkin headers, and we don't extract them from the Thrift request.

		log.Printf("Thrift (TCP) server on %s", *thriftTCPAddr)
		errc <- thrift.NewTSimpleServer4(
			thriftadd.NewAddServiceProcessor(thriftBinding{ctx, e}),
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
