package service

import (
	"context"
	"errors"
)

type Service struct {
}

func (s Service) Foo(ctx context.Context, i int, string1 string) (int, error) {
	panic(errors.New("not implemented"))
}
