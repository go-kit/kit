package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/peterbourgon/gokit/server"
	"golang.org/x/net/context"
)

func main() {
	// Parameterize our service.
	httpAddr := flag.String("http.addr", ":8080", "HTTP listen address")
	flag.Parse()

	// Business (service) domain setup.
	var a Add = add             // pure implementation
	a = logging(logWriter{}, a) // chained responsibilities

	// Mechanical stuff.
	root := context.Background()
	errc := make(chan error, 2)

	// Handle interactive interrupts.
	go func() {
		errc <- interrupt()
	}()

	// Bind a codec+endpoint to a listening point.
	// (We could support many of these in one process.)
	go func() {
		ctx, cancel := context.WithCancel(root)
		defer cancel()
		codec := jsonCodec{}
		endpoint := makeEndpoint(a)
		http.Handle("/add", server.HTTPEndpoint(ctx, codec, endpoint))
		log.Printf("HTTP server listening on %s", *httpAddr)
		errc <- http.ListenAndServe(*httpAddr, nil)
	}()

	log.Fatal(<-errc)
}

func interrupt() error {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	return fmt.Errorf("%v", <-c)
}
