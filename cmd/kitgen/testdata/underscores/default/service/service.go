package service

import "context"

import "errors"

type stubService struct {
}

func (s stubService) Foo(ctx context.Context, i int) (int, error) {
	panic(errors.New("not implemented"))
}
