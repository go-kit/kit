package service

import (
	"context"
	"errors"
)

type FooService struct {
}

func (f FooService) Bar(ctx context.Context, i int, s string) (string, error) {
	panic(errors.New("not implemented"))
}
