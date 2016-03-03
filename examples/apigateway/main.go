package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/context"

	"github.com/gorilla/mux"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/loadbalancer"
	"github.com/go-kit/kit/loadbalancer/consul"
	klog "github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"

	"github.com/hashicorp/consul/api"
)

var (
	discoveryClient consul.Client
	ctx             = context.Background()
)

func main() {

	consulConfig := api.DefaultConfig()
	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		log.Fatal(err)
	}
	discoveryClient = consul.NewClient(consulClient)

	r := mux.NewRouter()
	r.HandleFunc("/api/{service}/{method}", apiGateway)

	http.ListenAndServe(":8000", r)
}

func apiGateway(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	service := vars["service"]
	method := vars["method"]
	e, err := getEndpoint(service, method)
	if err != nil {
		log.Print(err)
		return
	}

	var val interface{}
	dec := json.NewDecoder(r.Body)
	err = dec.Decode(&val)
	if err != nil {
		log.Print(err)
		return
	}

	resp, err := e(ctx, val)
	if err != nil {
		log.Print(err)
		return
	}
	enc := json.NewEncoder(w)
	err = enc.Encode(resp)
	if err != nil {
		log.Print(err)
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

	publisher, err := consul.NewPublisher(discoveryClient, factory(ctx, method), klog.NewLogfmtLogger(&klog.StdlibWriter{}), se)
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
	log.Printf("encode req: %v", request)
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(request); err != nil {
		log.Print(err)
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
