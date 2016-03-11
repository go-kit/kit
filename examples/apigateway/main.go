package main

import (
	"encoding/json"
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
	logger := log.NewLogfmtLogger(os.Stderr)
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
	discoveryClient := consul.NewClient(consulClient)

	ctx := context.Background()

	// service definitions
	serviceDefs := []*ServiceDef{}
	serviceDefs = append(serviceDefs, &ServiceDef{
		Name: "addsvc",
		Endpoints: map[string]loadbalancer.Factory{
			"/api/addsvc/concat": factoryAddsvc(ctx, logger, makeConcatEndpoint),
			"/api/addsvc/sum":    factoryAddsvc(ctx, logger, makeSumEndpoint),
		},
	})
	serviceDefs = append(serviceDefs, &ServiceDef{
		Name: "stringsvc",
		Endpoints: map[string]loadbalancer.Factory{
			"/api/stringsvc/uppercase": routeFactory(ctx, "uppercase"),
			"/api/stringsvc/count":     routeFactory(ctx, "count"),
		},
	})

	// discover instances and register endpoints
	r := mux.NewRouter()
	for _, def := range serviceDefs {
		for path, e := range def.Endpoints {
			pub, err := consul.NewPublisher(discoveryClient, e, logger, def.Name)
			if err != nil {
				logger.Log("fatal", err)
			}
			r.HandleFunc(path, makeHandler(ctx, loadbalancer.NewRoundRobin(pub), logger))
		}
	}

	// apigateway
	go func() {
		errc <- http.ListenAndServe(*httpAddr, r)
	}()

	// wait for interrupt/error
	logger.Log("fatal", <-errc)
}

type ServiceDef struct {
	Name      string
	Endpoints map[string]loadbalancer.Factory
}

func interrupt() error {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	return fmt.Errorf("%s", <-c)
}

func makeHandler(ctx context.Context, lb loadbalancer.LoadBalancer, logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		e, err := lb.Endpoint()
		if err != nil {
			logger.Log("error", err)
			return
		}
		resp, err := e(ctx, r.Body)
		if err != nil {
			logger.Log("error", err)
			return
		}
		b, ok := resp.([]byte)
		if !ok {
			logger.Log("error", "endpoint response is not of type []byte")
			return
		}
		_, err = w.Write(b)
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
		r := request.(io.Reader)
		var req server.SumRequest
		if err := json.NewDecoder(r).Decode(&req); err != nil {
			return nil, err
		}
		v := svc.Sum(req.A, req.B)
		return json.Marshal(v)
	}
}

func makeConcatEndpoint(svc server.AddService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		r := request.(io.Reader)
		var req server.ConcatRequest
		if err := json.NewDecoder(r).Decode(&req); err != nil {
			return nil, err
		}
		v := svc.Concat(req.A, req.B)
		return json.Marshal(v)
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
