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
	"time"

	"github.com/gorilla/mux"
	"github.com/hashicorp/consul/api"
	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/examples/addsvc/client/grpc"
	"github.com/go-kit/kit/examples/addsvc/server"
	"github.com/go-kit/kit/loadbalancer"
	"github.com/go-kit/kit/loadbalancer/consul"
	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
)

func main() {
	var (
		httpAddr     = flag.String("http.addr", ":8000", "Address for HTTP (JSON) server")
		consulAddr   = flag.String("consul.addr", "", "Consul agent address")
		retryMax     = flag.Int("retry.max", 3, "per-request retries to different instances")
		retryTimeout = flag.Duration("retry.timeout", 500*time.Millisecond, "per-request timeout, including retries")
	)
	flag.Parse()

	// Log domain
	logger := log.NewLogfmtLogger(os.Stderr)
	logger = log.NewContext(logger).With("ts", log.DefaultTimestampUTC).With("caller", log.DefaultCaller)
	stdlog.SetFlags(0)                             // flags are handled by Go kit's logger
	stdlog.SetOutput(log.NewStdlibAdapter(logger)) // redirect anything using stdlib log to us

	// Service discovery domain. In this example we use Consul.
	consulConfig := api.DefaultConfig()
	if len(*consulAddr) > 0 {
		consulConfig.Address = *consulAddr
	}
	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		logger.Log("err", err)
		os.Exit(1)
	}
	discoveryClient := consul.NewClient(consulClient)

	// Context domain.
	ctx := context.Background()

	// Set up our routes.
	//
	// Each Consul service name maps to multiple instances of that service. We
	// connect to each instance according to its pre-determined transport: in this
	// case, we choose to access addsvc via its gRPC client, and stringsvc over
	// plain transport/http (it has no client package).
	//
	// Each service instance implements multiple methods, and we want to map each
	// method to a unique path on the API gateway. So, we define that path and its
	// corresponding factory function, which takes an instance string and returns an
	// endpoint.Endpoint for the specific method.
	//
	// Finally, we mount that path + endpoint handler into the router.
	r := mux.NewRouter()
	for consulName, methods := range map[string][]struct {
		path    string
		factory loadbalancer.Factory
	}{
		"addsvc": {
			{path: "/api/addsvc/concat", factory: grpc.ConcatEndpointFactory},
			{path: "/api/addsvc/sum", factory: grpc.SumEndpointFactory},
		},
		"stringsvc": {
			{path: "/api/stringsvc/uppercase", factory: httpFactory(ctx, "GET", "uppercase/")},
			{path: "/api/stringsvc/concat", factory: httpFactory(ctx, "GET", "concat/")},
		},
	} {
		for _, method := range methods {
			publisher, err := consul.NewPublisher(discoveryClient, method.factory, logger, consulName)
			if err != nil {
				logger.Log("service", consulName, "path", method.path, "err", err)
				continue
			}
			lb := loadbalancer.NewRoundRobin(publisher)
			e := loadbalancer.Retry(*retryMax, *retryTimeout, lb)
			h := makeHandler(ctx, e, logger)
			r.HandleFunc(method.path, h)
		}
	}

	// Mechanical stuff.
	errc := make(chan error)
	go func() {
		errc <- interrupt()
	}()
	go func() {
		logger.Log("transport", "http", "addr", *httpAddr)
		errc <- http.ListenAndServe(*httpAddr, r)
	}()
	logger.Log("err", <-errc)
}

func makeHandler(ctx context.Context, e endpoint.Endpoint, logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := e(ctx, r.Body)
		if err != nil {
			logger.Log("err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		b, ok := resp.([]byte)
		if !ok {
			logger.Log("err", "endpoint response is not of type []byte")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = w.Write(b)
		if err != nil {
			logger.Log("err", err)
			return
		}
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

func httpFactory(ctx context.Context, method, path string) loadbalancer.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		var e endpoint.Endpoint
		if !strings.HasPrefix(instance, "http") {
			instance = "http://" + instance
		}
		u, err := url.Parse(instance)
		if err != nil {
			return nil, nil, err
		}
		u.Path = path

		e = httptransport.NewClient(method, u, passEncode, passDecode).Endpoint()
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

func interrupt() error {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	return fmt.Errorf("%s", <-c)
}
