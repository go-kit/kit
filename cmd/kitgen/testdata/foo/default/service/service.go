package service

import "context"

import "errors"

type stubFooService struct {
}

func (f stubFooService) Bar(ctx context.Context, i int, s string) (string, error) {
	panic(errors.New("not implemented"))
}
