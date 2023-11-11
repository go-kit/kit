package endpoint_test

import (
	"context"
	"fmt"

	"github.com/openmesh/kit/endpoint"
)

func ExampleChain() {
	e := endpoint.Chain(
		annotate("first"),
		annotate("second"),
		annotate("third"),
	)(myEndpoint)

	if _, err := e(ctx, req); err != nil {
		panic(err)
	}

	// Output:
	// first pre
	// second pre
	// third pre
	// my endpoint!
	// third post
	// second post
	// first post
}

var (
	ctx = context.Background()
	req = struct{}{}
)

func annotate(s string) endpoint.Middleware[interface{}, interface{}] {
	return func(next endpoint.Endpoint[interface{}, interface{}]) endpoint.Endpoint[interface{}, interface{}] {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			fmt.Println(s, "pre")
			defer fmt.Println(s, "post")
			return next(ctx, request)
		}
	}
}

func myEndpoint(context.Context, interface{}) (interface{}, error) {
	fmt.Println("my endpoint!")
	return struct{}{}, nil
}
