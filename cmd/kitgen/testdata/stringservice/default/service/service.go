package service

import "golang.org/x/net/context"
import "context"

import "errors"

type stubService struct {
}

func (s stubService) Concat(ctx context.Context, a string, b string) (string, error) {
	panic(errors.New("not implemented"))
}
func (s stubService) Count(ctx context.Context, string1 string) int {
	panic(errors.New("not implemented"))
}
