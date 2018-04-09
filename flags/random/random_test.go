// Package random provides feature flags that will return one response from
// a provided discrete list of options.
package random

import (
	"context"
	crand "crypto/rand"
	"fmt"
	"io"
	"math/rand"
	"testing"
	"time"
)

func TestRandomBooler(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	hasTrue := false
	hasFalse := false

	alwaysTrue := NewBooler(r, true)
	alwaysFalse := NewBooler(r, false)
	equal := NewBooler(r, true, false)

	attempts := 16
	for i := 0; i < attempts; i++ {
		if !alwaysTrue.Bool(context.Background()) {
			t.Errorf("should always be true")
		}
		if alwaysFalse.Bool(context.Background()) {
			t.Errorf("should always be false")
		}

		if equal.Bool(context.Background()) {
			hasTrue = true
		} else {
			hasFalse = true
		}
		if hasTrue && hasFalse {
			return
		}
	}
	if hasTrue {
		t.Errorf("never false in %d tries", attempts)
	} else {
		t.Errorf("never true in %d tries", attempts)
	}
}

func TestRandomInter(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	opts := []int64{}
	expected := map[int64]bool{}
	for i := 0; i < 10; i++ {
		val := r.Int63()
		opts = append(opts, val)
		expected[val] = true
	}

	inter := NewInter(r, opts...)

	for i := 0; i < 100; i++ {
		val := inter.Int(context.Background())
		if found := expected[val]; !found {
			t.Errorf("%d not found in %v\n", val, expected)
		}
	}
}

func TestRandomFloater(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	opts := []float64{}
	expected := map[float64]bool{}
	for i := 0; i < 10; i++ {
		val := float64(r.Int63() / 17.0)
		opts = append(opts, val)
		expected[val] = true
	}

	floater := NewFloater(r, opts...)

	for i := 0; i < 100; i++ {
		val := floater.Float(context.Background())
		if found := expected[val]; !found {
			t.Errorf("%f not found in %v\n", val, expected)
		}
	}
}

func TestRandomStringer(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	opts := []string{}
	expected := map[string]bool{}
	for i := 0; i < 10; i++ {
		buf := make([]byte, 1+i)
		io.ReadFull(crand.Reader, buf)
		val := fmt.Sprintf("%x", buf)
		opts = append(opts, val)
		expected[val] = true
	}

	stringer := NewStringer(r, opts...)

	for i := 0; i < 100; i++ {
		val := stringer.String(context.Background())
		if found := expected[val]; !found {
			t.Errorf("%q not found in %v\n", val, expected)
		}
	}
}
