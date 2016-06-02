package service_test

import (
	"fmt"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/service"
)

func ExampleChain() {
	s := service.Chain(
		annotate("first"),
		annotate("second"),
		annotate("third"),
	)(service.Func(myServiceFunc))

	if _, err := s.Endpoint("foobar"); err != nil {
		panic(err)
	}

	// Output:
	// first pre foobar
	// second pre foobar
	// third pre foobar
	// my service func foobar
	// third post foobar
	// second post foobar
	// first post foobar
}

func annotate(s string) service.Middleware {
	return func(next service.Service) service.Service {
		return service.Func(func(method string) (endpoint.Endpoint, error) {
			fmt.Println(s, "pre", method)
			defer fmt.Println(s, "post", method)
			return next.Endpoint(method)
		})
	}
}

func myServiceFunc(method string) (endpoint.Endpoint, error) {
	fmt.Println("my service func", method)
	return endpoint.Nop, nil
}
