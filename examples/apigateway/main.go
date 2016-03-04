package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	stdlog "log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/go-kit/kit/endpoint"
	addsvc "github.com/go-kit/kit/examples/addsvc/client/grpc"
	"github.com/go-kit/kit/examples/addsvc/server"
	"github.com/go-kit/kit/loadbalancer"
	"github.com/go-kit/kit/loadbalancer/consul"
	log "github.com/go-kit/kit/log"
	//grpctransport "github.com/go-kit/kit/transport/grpc"
	httptransport "github.com/go-kit/kit/transport/http"
	//proto "github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"
	"github.com/hashicorp/consul/api"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	discoveryClient consul.Client
	ctx             = context.Background()
	logger          log.Logger
)

func main() {
	fs := flag.NewFlagSet("", flag.ExitOnError)
	var (
		httpAddr   = fs.String("http.addr", ":8000", "Address for HTTP (JSON) server")
		consulAddr = fs.String("consul.addr", "", "Consul agent address")
	)
	flag.Usage = fs.Usage
	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}

	// log
	logger = log.NewLogfmtLogger(os.Stderr)
	logger = log.NewContext(logger).With("ts", log.DefaultTimestampUTC).With("caller", log.DefaultCaller)
	stdlog.SetFlags(0)                             // flags are handled by Go kit's logger
	stdlog.SetOutput(log.NewStdlibAdapter(logger)) // redirect anything using stdlib log to us

	// errors
	errc := make(chan error)
	go func() {
		errc <- interrupt()
	}()

	// consul
	consulConfig := api.DefaultConfig()
	if len(*consulAddr) > 0 {
		consulConfig.Address = *consulAddr
	}
	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		logger.Log("fatal", err)
	}
	discoveryClient = consul.NewClient(consulClient)

	// discover service stringsvc
	uppercase, err := consul.NewPublisher(discoveryClient, routeFactory(ctx, "uppercase"), logger, "stringsvc")
	if err != nil {
		logger.Log("fatal", err)
	}
	count, err := consul.NewPublisher(discoveryClient, routeFactory(ctx, "count"), logger, "stringsvc")
	if err != nil {
		logger.Log("fatal", err)
	}

	// discover service addsvc
	addsvcSum, err := consul.NewPublisher(discoveryClient, factoryAddsvc(ctx, logger, makeSumEndpoint), logger, "addsvc")
	if err != nil {
		logger.Log("fatal", err)
	}
	addsvcConcat, err := consul.NewPublisher(discoveryClient, factoryAddsvc(ctx, logger, makeConcatEndpoint), logger, "addsvc")
	if err != nil {
		logger.Log("fatal", err)
	}

	// apigateway
	go func() {
		r := mux.NewRouter()
		r.HandleFunc("/api/addsvc/sum", makeSumHandler(ctx, loadbalancer.NewRoundRobin(addsvcSum)))
		r.HandleFunc("/api/addsvc/concat", makeConcatHandler(ctx, loadbalancer.NewRoundRobin(addsvcConcat)))
		r.HandleFunc("/api/stringsvc/uppercase", factoryPassHandler(loadbalancer.NewRoundRobin(uppercase), logger))
		r.HandleFunc("/api/stringsvc/count", factoryPassHandler(loadbalancer.NewRoundRobin(count), logger))
		errc <- http.ListenAndServe(*httpAddr, r)
	}()

	// wait for interrupt/error
	logger.Log("fatal", <-errc)
}

func interrupt() error {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	return fmt.Errorf("%s", <-c)
}

func makeSumHandler(ctx context.Context, lb loadbalancer.LoadBalancer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sumReq, err := server.DecodeSumRequest(r)
		if err != nil {
			logger.Log("error", err)
			return
		}
		e, err := lb.Endpoint()
		if err != nil {
			logger.Log("error", err)
			return
		}
		sumResp, err := e(ctx, sumReq)
		if err != nil {
			logger.Log("error", err)
			return
		}
		err = server.EncodeSumResponse(w, sumResp)
		if err != nil {
			logger.Log("error", err)
			return
		}
	}
}

func makeConcatHandler(ctx context.Context, lb loadbalancer.LoadBalancer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		concatReq, err := server.DecodeConcatRequest(r)
		if err != nil {
			logger.Log("error", err)
			return
		}
		e, err := lb.Endpoint()
		if err != nil {
			logger.Log("error", err)
			return
		}
		concatResp, err := e(ctx, concatReq)
		if err != nil {
			logger.Log("error", err)
			return
		}
		err = server.EncodeConcatResponse(w, concatResp)
		if err != nil {
			logger.Log("error", err)
			return
		}
	}
}

func factoryAddsvc(ctx context.Context, logger log.Logger, maker func(server.AddService) endpoint.Endpoint) loadbalancer.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		var e endpoint.Endpoint
		conn, err := grpc.Dial(instance, grpc.WithInsecure())
		if err != nil {
			return e, nil, err
		}
		svc := addsvc.New(ctx, conn, logger)
		return maker(svc), nil, nil
	}
}

func makeSumEndpoint(svc server.AddService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(server.SumRequest)
		v := svc.Sum(req.A, req.B)
		return server.SumResponse{V: v}, nil
	}
}

func makeConcatEndpoint(svc server.AddService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(server.ConcatRequest)
		v := svc.Concat(req.A, req.B)
		return server.ConcatResponse{V: v}, nil
	}
}

func routeFactory(ctx context.Context, method string) loadbalancer.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		var e endpoint.Endpoint
		if !strings.HasPrefix(instance, "http") {
			instance = "http://" + instance
		}
		u, err := url.Parse(instance)
		if err != nil {
			return nil, nil, err
		}
		u.Path = method

		e = httptransport.NewClient("GET", u, passEncode, passDecode).Endpoint()
		return e, nil, nil
	}
}

func passEncode(r *http.Request, request interface{}) error {
	r.Body = request.(io.ReadCloser)
	return nil
}

func passDecode(r *http.Response) (interface{}, error) {
	return ioutil.ReadAll(r.Body)
}

func factoryPassHandler(lb loadbalancer.LoadBalancer, logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		e, err := lb.Endpoint()
		if err != nil {
			logger.Log("error", err)
			return
		}
		resp, err := e(ctx, r.Body)
		if err != nil {
			logger.Log("warning", err)
			fmt.Fprint(w, err)
			return
		}
		b := resp.([]byte)
		_, err = w.Write(b)
		if err != nil {
			logger.Log("warning", err)
			fmt.Fprint(w, err)
			return
		}
	}
}
