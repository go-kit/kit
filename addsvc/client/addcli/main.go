package main

import (
	"flag"
	"log"
	"net/rpc"
	"os"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	grpcclient "github.com/go-kit/kit/addsvc/client/grpc"
	netrpcclient "github.com/go-kit/kit/addsvc/client/netrpc"
	"github.com/go-kit/kit/addsvc/reqrep"
	"github.com/go-kit/kit/endpoint"
)

func main() {
	// Flag domain. Note that gRPC transitively registers flags via its import
	// of glog. So, we define a new flag set, to keep those domains distinct.
	fs := flag.NewFlagSet("", flag.ExitOnError)
	var (
		transport  = fs.String("transport", "grpc", "grpc, netrpc")
		grpcAddr   = fs.String("grpc.addr", "localhost:8002", "gRPC address")
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
	case "grpc":
		cc, err := grpc.Dial(*grpcAddr)
		if err != nil {
			log.Fatalf("grpc.Dial: %v", err)
		}
		e = grpcclient.NewClient(cc)

	case "netrpc":
		client, err := rpc.DialHTTP("tcp", *netrpcAddr)
		if err != nil {
			log.Fatalf("rpc.DialHTTP: %v", err)
		}
		e = netrpcclient.NewClient(client)

	default:
		log.Fatalf("unsupported transport %q", *transport)
	}

	response, err := e(context.Background(), reqrep.AddRequest{A: *a, B: *b})
	if err != nil {
		log.Fatalf("when invoking request: %v", err)
	}
	addResponse, ok := response.(reqrep.AddResponse)
	if !ok {
		log.Fatalf("when type-asserting response: %v", endpoint.ErrBadCast)
	}
	log.Print(addResponse.V)
}
