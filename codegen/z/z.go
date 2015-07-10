package z

import (
	"fmt"

	"golang.org/x/net/context"
)

type X interface {
	Y(context.Context, struct {
		F fmt.Stringer
		G uint32
	}, int, int) int64
	// Z(ctx context.Context, a, b int) (r int, err error)
}
