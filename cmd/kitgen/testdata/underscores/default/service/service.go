package service

import (
	"context"
	"errors"
)

type Service struct {
}

func (s Service) Foo(ctx context.Context, i int) (int, error) {
	panic(errors.New("not implemented"))
}
