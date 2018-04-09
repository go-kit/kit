// Package random provides feature flags that will return one response from
// a provided discrete list of options.
package random

import (
	"context"
	"math/rand"

	"github.com/go-kit/kit/flags"
)

// NewBooler builds a Booler that returns one of a discrete set of options
func NewBooler(r *rand.Rand, opts ...bool) flags.Booler {
	return flags.BoolerFunc(func(context.Context) bool {
		return opts[r.Intn(len(opts))]
	})
}

// NewInter builds an Inter that returns one of a discrete set of options
func NewInter(r *rand.Rand, opts ...int64) flags.Inter {
	return flags.InterFunc(func(context.Context) int64 {
		return opts[r.Intn(len(opts))]
	})
}

// NewFloater builds a Floater that returns one of a discrete set of options
func NewFloater(r *rand.Rand, opts ...float64) flags.Floater {
	return flags.FloaterFunc(func(context.Context) float64 {
		return opts[r.Intn(len(opts))]
	})
}

// NewStringer builds a Stringer that returns one of a discrete set of options
func NewStringer(r *rand.Rand, opts ...string) flags.Stringer {
	return flags.StringerFunc(func(context.Context) string {
		return opts[r.Intn(len(opts))]
	})
}
