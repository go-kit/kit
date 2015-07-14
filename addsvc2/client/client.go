package main

import (
	"flag"
	"log"
	"net/rpc"
	"net/url"
	"os"
	"strings"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/sasha-s/kit/addsvc2/add"
)

func main() {
	// Flag domain. Note that gRPC transitively registers flags via its import
	// of glog. So, we define a new flag set, to keep those domains distinct.
	fs := flag.NewFlagSet("", flag.ExitOnError)
	var (
		transport  = fs.String("transport", "http", "http, netrpc")
		httpAddr   = fs.String("http.addr", "localhost:8001", "HTTP (JSON) address")
		netrpcAddr = fs.String("netrpc.addr", "localhost:8003", "net/rpc address")
		a          = fs.Int64("a", 1, "a value")
		b          = fs.Int64("b", 2, "b value")
	)
	flag.Usage = fs.Usage // only show our flags
	fs.Parse(os.Args[1:])
	log.SetFlags(0)
	log.SetOutput(os.Stdout)

	var e endpoint.Endpoint
	switch *transport {
	case "http":
		if !strings.HasPrefix(*httpAddr, "http") {
			*httpAddr = "http://" + *httpAddr
		}
		u, err := url.Parse(*httpAddr)
		if err != nil {
			log.Fatalf("url.Parse: %v", err)
		}
		if u.Path == "" {
			u.Path = "/add"
		}
		e = add.NewAdderAddHTTPClient("GET", u.String())

	case "netrpc":
		client, err := rpc.DialHTTP("tcp", *netrpcAddr)
		if err != nil {
			log.Fatalf("rpc.DialHTTP: %v", err)
		}
		e = add.NewAdderAddRPCClient(client)("Add.Add")

	default:
		log.Fatalf("unsupported transport %q", *transport)
	}

	response, err := e(context.Background(), add.AdderAddRequest{A: *a, B: *b})
	if err != nil {
		log.Fatalf("when invoking request: %v", err)
	}
	addResponse, ok := response.(add.AdderAddResponse)
	if !ok {
		log.Fatalf("when type-asserting response: %T", response)
	}
	log.Print(addResponse.V)
}
