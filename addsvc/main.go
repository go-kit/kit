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

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/peterbourgon/gokit/addsvc/pb"
	"github.com/peterbourgon/gokit/metrics"
	"github.com/peterbourgon/gokit/metrics/expvar"
	"github.com/peterbourgon/gokit/metrics/statsd"
	"github.com/peterbourgon/gokit/server"
	"github.com/peterbourgon/gokit/server/zipkin"
	"github.com/peterbourgon/gokit/transport/http/cors"
)

func main() {
	var (
		httpJSONAddr = flag.String("http.json.addr", ":8001", "Address for HTTP/JSON server")
		grpcTCPAddr  = flag.String("grpc.tcp.addr", ":8002", "Address for gRPC (TCP) server")
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
