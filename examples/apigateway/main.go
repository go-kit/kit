package main

import (
	"bytes"
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

	"golang.org/x/net/context"

	"github.com/gorilla/mux"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/loadbalancer"
	"github.com/go-kit/kit/loadbalancer/consul"
	log "github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"

	"github.com/hashicorp/consul/api"
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

	// apigateway
	go func() {
		r := mux.NewRouter()
		r.HandleFunc("/api/{service}/{method}", apiGateway)
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

func apiGateway(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	service := vars["service"]
	method := vars["method"]
	e, err := getEndpoint(service, method)
	if err != nil {
		logger.Log("error", err)
		return
	}

	var val interface{}
	dec := json.NewDecoder(r.Body)
	err = dec.Decode(&val)
	if err != nil {
		logger.Log("warning", err)
		fmt.Fprint(w, err)
		return
	}

	resp, err := e(ctx, val)
	if err != nil {
		logger.Log("warning", err)
		fmt.Fprint(w, err)
		return
	}
	enc := json.NewEncoder(w)
	err = enc.Encode(resp)
	if err != nil {
		logger.Log("warning", err)
		fmt.Fprint(w, err)
		return
	}
}

var services = make(map[string]service)

type service map[string]loadbalancer.LoadBalancer

func getEndpoint(se string, method string) (endpoint.Endpoint, error) {
	if s, ok := services[se]; ok {
		if m, ok := s[method]; ok {
			return m.Endpoint()
		}
	}

	publisher, err := consul.NewPublisher(discoveryClient, factory(ctx, method), log.NewLogfmtLogger(&log.StdlibWriter{}), se)
	publisher.Endpoints()
	if err != nil {
		return nil, err
	}
	rr := loadbalancer.NewRoundRobin(publisher)

	if _, ok := services[se]; ok {
		services[se][method] = rr
	} else {
		services[se] = service{method: rr}
	}

	return rr.Endpoint()
}

func factory(ctx context.Context, method string) loadbalancer.Factory {
	return func(service string) (endpoint.Endpoint, io.Closer, error) {
		var e endpoint.Endpoint
		e = makeProxy(ctx, service, method)
		return e, nil, nil
	}
}

func makeProxy(ctx context.Context, service, method string) endpoint.Endpoint {
	if !strings.HasPrefix(service, "http") {
		service = "http://" + service
	}
	u, err := url.Parse(service)
	if err != nil {
		panic(err)
	}
	if u.Path == "" {
		u.Path = "/" + method
	}

	return httptransport.NewClient(
		"GET",
		u,
		encodeRequest,
		decodeResponse,
	).Endpoint()
}

func encodeRequest(r *http.Request, request interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(request); err != nil {
		return err
	}
	r.Body = ioutil.NopCloser(&buf)
	return nil
}

func decodeResponse(r *http.Response) (interface{}, error) {
	var response interface{}
	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		return nil, err
	}
	return response, nil
}
