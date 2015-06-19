package http

import (
	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
)

func NewClient() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		return nil, nil
	}
}
