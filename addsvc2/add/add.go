package add

import "golang.org/x/net/context"

//go:generate gokit-gen -package github.com/sasha-s/kit/addsvc2/add -type Adder -w -binding=rpc,http

type Adder interface {
	Add(ctx context.Context, a int64, b int64) (v int64)
}
