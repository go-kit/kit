package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/go-kit/kit/addsvc/reqrep"
	"github.com/go-kit/kit/endpoint"
	"golang.org/x/net/context"

	grpcclient "github.com/go-kit/kit/addsvc/client/grpc"

	"google.golang.org/grpc"
)

func main() {
	// Flag domain. Note that gRPC transitively registers flags via its import
	// of glog. So, we define a new flag set, to keep those domains distinct.
	fs := flag.NewFlagSet("", flag.ExitOnError)
	var (
		grpcAddr    = fs.String("grpc.addr", "localhost:8002", "gRPC address")
		grpcTimeout = fs.Duration("grpc.timeout", 250*time.Millisecond, "gRPC dial timeout")
		a           = fs.Int64("a", 1, "a value")
		b           = fs.Int64("b", 2, "b value")
	)
	flag.Usage = fs.Usage // only show our flags
	fs.Parse(os.Args[1:])
	log.SetFlags(0)
	log.SetOutput(os.Stdout)

	cc, err := grpc.Dial(*grpcAddr, grpc.WithTimeout(*grpcTimeout))
	if err != nil {
		log.Fatal(err)
	}

	var e endpoint.Endpoint = grpcclient.NewClient(cc)
	response, err := e(context.Background(), reqrep.AddRequest{A: *a, B: *b})
	if err != nil {
		log.Fatalf("request: %v", err)
	}

	addResponse, ok := response.(reqrep.AddResponse)
	if !ok {
		log.Fatalf("response: %v", endpoint.ErrBadCast)
	}

	log.Print(addResponse.V)
}
