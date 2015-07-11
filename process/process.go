//go:generate go run ../gen-cmd/main.go process
//go:generate go fmt
package process

import "golang.org/x/net/context"

type Process func(ctx context.Context, number int, head, tails string) (ret string, err error)

func pureProcess(ctx context.Context, number int, head, tails string) (ret string, err error) {
	return "test", nil
}
