package main

import (
	"fmt"
	"io"
	"log"
	"time"

	"golang.org/x/net/context"
)

// Add is a function that takes two ints and returns an int.
type Add func(context.Context, int, int) int

// add is a pure implementation of the Add definition.
// It can therefore ignore the context.
func add(_ context.Context, a, b int) int {
	return a + b
}

// addVia is a non-pure implementation of the Add definition, and
// therefore must thread the Context through to the backing Resource.
func addVia(r Resource) Add {
	return func(ctx context.Context, a, b int) int {
		return r.GetValue(ctx, a) + r.GetValue(ctx, b)
	}
}

// Resource represents some arbitrary backing resource, like another service.
type Resource interface {
	GetValue(context.Context, int) int
}

type resource struct{}

// GetA represents some transform of an integer value.
func (r resource) GetValue(_ context.Context, i int) int {
	return i
}

func logging(dst io.Writer, add Add) Add {
	return func(ctx context.Context, a, b int) (v int) {
		defer func(begin time.Time) {
			fmt.Fprintf(dst, "Add(%d, %d) = %d (%s)\n", a, b, v, time.Since(begin))
		}(time.Now())

		v = add(ctx, a, b)
		return
	}
}

type logWriter struct{}

func (w logWriter) Write(p []byte) (int, error) { log.Printf("%s", p); return len(p), nil }
