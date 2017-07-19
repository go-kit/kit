package foo

import "golang.org/x/net/context"

type Service interface {
	Concat(ctx context.Context, a, b string) (string, error)
	Count(ctx context.Context, s string) (count int)
}

