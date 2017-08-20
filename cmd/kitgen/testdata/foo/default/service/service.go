package service

import "context"
import "encoding/json"
import "errors"
import "net/http"
import "github.com/go-kit/kit/endpoint"
import httptransport "github.com/go-kit/kit/transport/http"

type stubFooService struct {
}

func (f stubFooService) Bar(ctx context.Context, i int, s string) (string, error) {
	panic(errors.New("not implemented"))
}
