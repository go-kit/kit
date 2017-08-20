package service

import "context"
import "encoding/json"
import "errors"
import "net/http"
import "github.com/go-kit/kit/endpoint"
import httptransport "github.com/go-kit/kit/transport/http"

type stubService struct {
}

func (s stubService) Foo(ctx context.Context, i int) (int, error) {
	panic(errors.New("not implemented"))
}
