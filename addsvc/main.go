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

	"github.com/streadway/handy/cors"
	"github.com/streadway/handy/encoding"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

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
		httpAddr = flag.String("http.addr", ":8001", "Address for HTTP (JSON) server")
		grpcAddr = flag.String("grpc.addr", ":8002", "Address for gRPC server")
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
		before := kithttp.Before(zipkin.ToContext(zipkin.ViaHTTP(zipkinSpanFunc)))
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
