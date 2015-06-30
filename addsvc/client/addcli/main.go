package main

import (
	"flag"
	"log"
	"net/rpc"
	"os"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/apache/thrift/lib/go/thrift"
	thriftadd "github.com/go-kit/kit/addsvc/_thrift/gen-go/add"
	grpcclient "github.com/go-kit/kit/addsvc/client/grpc"
	netrpcclient "github.com/go-kit/kit/addsvc/client/netrpc"
	thriftclient "github.com/go-kit/kit/addsvc/client/thrift"
	"github.com/go-kit/kit/addsvc/reqrep"
	"github.com/go-kit/kit/endpoint"
)

func main() {
	// Flag domain. Note that gRPC transitively registers flags via its import
	// of glog. So, we define a new flag set, to keep those domains distinct.
	fs := flag.NewFlagSet("", flag.ExitOnError)
	var (
		transport        = fs.String("transport", "grpc", "grpc, netrpc, thrift")
		grpcAddr         = fs.String("grpc.addr", "localhost:8002", "gRPC address")
		netrpcAddr       = fs.String("netrpc.addr", "localhost:8003", "net/rpc address")
		thriftAddr       = fs.String("thrift.addr", "localhost:8004", "Thrift address")
		thriftProtocol   = fs.String("thrift.protocol", "binary", "binary, compact, json, simplejson")
		thriftBufferSize = fs.Int("thrift.buffer.size", 0, "0 for unbuffered")
		thriftFramed     = fs.Bool("thrift.framed", false, "true to enable framing")
		a                = fs.Int64("a", 1, "a value")
		b                = fs.Int64("b", 2, "b value")
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

	case "thrift":
		var protocolFactory thrift.TProtocolFactory
		switch *thriftProtocol {
		case "compact":
			protocolFactory = thrift.NewTCompactProtocolFactory()
		case "simplejson":
			protocolFactory = thrift.NewTSimpleJSONProtocolFactory()
		case "json":
			protocolFactory = thrift.NewTJSONProtocolFactory()
		case "binary", "":
			protocolFactory = thrift.NewTBinaryProtocolFactoryDefault()
		default:
			log.Fatalf("invalid protocol %q", *thriftProtocol)
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

		transportSocket, err := thrift.NewTSocket(*thriftAddr)
		if err != nil {
			log.Fatalf("thrift.NewTSocket: %v", err)
		}

		transport := transportFactory.GetTransport(transportSocket)
		defer transport.Close()
		if err := transport.Open(); err != nil {
			log.Fatalf("Thrift transport.Open: %v", err)
		}

		e = thriftclient.NewClient(thriftadd.NewAddServiceClientFactory(transport, protocolFactory))

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
