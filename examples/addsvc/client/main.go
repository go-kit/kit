package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	grpcclient "github.com/go-kit/kit/examples/addsvc/client/grpc"
	httpjsonclient "github.com/go-kit/kit/examples/addsvc/client/httpjson"
	netrpcclient "github.com/go-kit/kit/examples/addsvc/client/netrpc"
	thriftclient "github.com/go-kit/kit/examples/addsvc/client/thrift"
	"github.com/go-kit/kit/loadbalancer"
	"github.com/go-kit/kit/loadbalancer/static"
	"github.com/go-kit/kit/log"
)

func main() {
	var (
		transport        = flag.String("transport", "httpjson", "httpjson, grpc, netrpc, thrift")
		httpAddrs        = flag.String("http.addrs", "localhost:8001", "Comma-separated list of addresses for HTTP (JSON) servers")
		grpcAddrs        = flag.String("grpc.addrs", "localhost:8002", "Comma-separated list of addresses for gRPC servers")
		netrpcAddrs      = flag.String("netrpc.addrs", "localhost:8003", "Comma-separated list of addresses for net/rpc servers")
		thriftAddrs      = flag.String("thrift.addrs", "localhost:8004", "Comma-separated list of addresses for Thrift servers")
		thriftProtocol   = flag.String("thrift.protocol", "binary", "binary, compact, json, simplejson")
		thriftBufferSize = flag.Int("thrift.buffer.size", 0, "0 for unbuffered")
		thriftFramed     = flag.Bool("thrift.framed", false, "true to enable framing")
	)
	flag.Parse()
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "\n%s [flags] method arg1 arg2\n\n", filepath.Base(os.Args[0]))
		flag.Usage()
		os.Exit(1)
	}

	randomSeed := time.Now().UnixNano()

	root := context.Background()
	method, s1, s2 := flag.Arg(0), flag.Arg(1), flag.Arg(2)

	var logger log.Logger
	logger = log.NewLogfmtLogger(os.Stdout)
	logger = log.NewContext(logger).With("caller", log.DefaultCaller)
	logger = log.NewContext(logger).With("transport", *transport)

	var (
		instances                 []string
		sumFactory, concatFactory loadbalancer.Factory
	)

	switch *transport {
	case "grpc":
		instances = strings.Split(*grpcAddrs, ",")
		sumFactory = grpcclient.SumEndpointFactory
		concatFactory = grpcclient.ConcatEndpointFactory

	case "httpjson":
		instances = strings.Split(*httpAddrs, ",")
		for i, rawurl := range instances {
			if !strings.HasPrefix("http", rawurl) {
				instances[i] = "http://" + rawurl
			}
		}
		sumFactory = httpjsonclient.SumEndpointFactory
		concatFactory = httpjsonclient.ConcatEndpointFactory

	case "netrpc":
		instances = strings.Split(*netrpcAddrs, ",")
		sumFactory = netrpcclient.SumEndpointFactory
		concatFactory = netrpcclient.ConcatEndpointFactory

	case "thrift":
		instances = strings.Split(*thriftAddrs, ",")
		thriftClient := thriftclient.New(*thriftProtocol, *thriftBufferSize, *thriftFramed, logger)
		sumFactory = thriftClient.SumEndpoint
		concatFactory = thriftClient.ConcatEndpoint

	default:
		logger.Log("err", "invalid transport")
		os.Exit(1)
	}

	sum := buildEndpoint(instances, sumFactory, randomSeed, logger)
	concat := buildEndpoint(instances, concatFactory, randomSeed, logger)

	svc := newClient(root, sum, concat, logger)

	begin := time.Now()
	switch method {
	case "sum":
		a, _ := strconv.Atoi(s1)
		b, _ := strconv.Atoi(s2)
		v := svc.Sum(a, b)
		logger.Log("method", "sum", "a", a, "b", b, "v", v, "took", time.Since(begin))

	case "concat":
		a, b := s1, s2
		v := svc.Concat(a, b)
		logger.Log("method", "concat", "a", a, "b", b, "v", v, "took", time.Since(begin))

	default:
		logger.Log("err", "invalid method "+method)
		os.Exit(1)
	}
}

func buildEndpoint(instances []string, factory loadbalancer.Factory, seed int64, logger log.Logger) endpoint.Endpoint {
	publisher := static.NewPublisher(instances, factory, logger)
	random := loadbalancer.NewRandom(publisher, seed)
	return loadbalancer.Retry(10, 10*time.Second, random)
}
