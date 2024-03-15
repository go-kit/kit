package endpoint_test

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/endpoint"
)

func ExampleChain() {
	e := endpoint.Chain(
		annotate[any, any]("first"),
		annotate[any, any]("second"),
		annotate[any, any]("third"),
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

func annotate[Req any, Resp any](s string) endpoint.Middleware[Req, Resp] {
	return func(next endpoint.Endpoint[Req, Resp]) endpoint.Endpoint[Req, Resp] {
		return endpoint.Endpoint[Req, Resp](func(ctx context.Context, request Req) (Resp, error) {
			fmt.Println(s, "pre")
			defer fmt.Println(s, "post")
			return next(ctx, request)
		})
	}
}

func myEndpoint(context.Context, interface{}) (interface{}, error) {
	fmt.Println("my endpoint!")
	return struct{}{}, nil
}
