package main

import (
	"flag"
	"fmt"
	"net/rpc"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	grpcclient "github.com/go-kit/kit/examples/addsvc/client/grpc"
	httpjsonclient "github.com/go-kit/kit/examples/addsvc/client/httpjson"
	netrpcclient "github.com/go-kit/kit/examples/addsvc/client/netrpc"
	thriftclient "github.com/go-kit/kit/examples/addsvc/client/thrift"
	"github.com/go-kit/kit/examples/addsvc/server"
	thriftadd "github.com/go-kit/kit/examples/addsvc/thrift/gen-go/add"
	"github.com/go-kit/kit/log"
)

func main() {
	var (
		transport        = flag.String("transport", "httpjson", "httpjson, grpc, netrpc, thrift")
		httpAddr         = flag.String("http.addr", "localhost:8001", "Address for HTTP (JSON) server")
		grpcAddr         = flag.String("grpc.addr", "localhost:8002", "Address for gRPC server")
		netrpcAddr       = flag.String("netrpc.addr", "localhost:8003", "Address for net/rpc server")
		thriftAddr       = flag.String("thrift.addr", "localhost:8004", "Address for Thrift server")
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

	root := context.Background()
	method, s1, s2 := flag.Arg(0), flag.Arg(1), flag.Arg(2)

	var logger log.Logger
	logger = log.NewLogfmtLogger(os.Stdout)
	logger = log.NewContext(logger).With("caller", log.DefaultCaller)
	logger = log.NewContext(logger).With("transport", *transport)

	var svc server.AddService
	switch *transport {
	case "grpc":
		cc, err := grpc.Dial(*grpcAddr, grpc.WithInsecure())
		if err != nil {
			logger.Log("err", err)
			os.Exit(1)
		}
		defer cc.Close()
		svc = grpcclient.New(root, cc, logger)

	case "httpjson":
		rawurl := *httpAddr
		if !strings.HasPrefix("http", rawurl) {
			rawurl = "http://" + rawurl
		}
		baseurl, err := url.Parse(rawurl)
		if err != nil {
			logger.Log("err", err)
			os.Exit(1)
		}
		svc = httpjsonclient.New(root, baseurl, logger, nil)

	case "netrpc":
		cli, err := rpc.DialHTTP("tcp", *netrpcAddr)
		if err != nil {
			logger.Log("err", err)
			os.Exit(1)
		}
		defer cli.Close()
		svc = netrpcclient.New(cli, logger)

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
			logger.Log("protocol", *thriftProtocol, "err", "invalid protocol")
			os.Exit(1)
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
			logger.Log("during", "thrift.NewTSocket", "err", err)
			os.Exit(1)
		}
		trans := transportFactory.GetTransport(transportSocket)
		defer trans.Close()
		if err := trans.Open(); err != nil {
			logger.Log("during", "thrift transport.Open", "err", err)
			os.Exit(1)
		}
		cli := thriftadd.NewAddServiceClientFactory(trans, protocolFactory)
		svc = thriftclient.New(cli, logger)

	default:
		logger.Log("err", "invalid transport")
		os.Exit(1)
	}

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
