package term_test

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"strconv"
	"sync"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/term"
	"gopkg.in/logfmt.v0"
)

type mymap map[int]int

func (m mymap) String() string { return "special_behavior" }

func TestColorLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := newColorLogger(t, &buf)

	if err := logger.Log("hello", "world"); err != nil {
		t.Fatal(err)
	}
	if want, have := "hello=world\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	buf.Reset()
	if err := logger.Log("a", 1, "err", errors.New("error")); err != nil {
		t.Fatal(err)
	}
	if want, have := "\u001b[32m\u001b[48ma=1 err=error\n\u001b[0m", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	buf.Reset()
	if err := logger.Log("std_map", map[int]int{1: 2}, "my_map", mymap{0: 0}); err != nil {
		t.Fatal(err)
	}
	if want, have := "std_map=\""+logfmt.ErrUnsupportedValueType.Error()+"\" my_map=special_behavior\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}
}

func newColorLogger(t testing.TB, w io.Writer) log.Logger {
	return term.NewColorLogger(log.NewLogfmtLogger(w),
		func(keyvals ...interface{}) term.FgBgColor {
			for i := 0; i < len(keyvals); i += 2 {
				key := term.AsString(keyvals[i])
				if key == "a" {
					return term.FgBgColor{Fg: term.Green, Bg: term.Default}
				}
				if key == "err" && keyvals[i+1] != nil {
					return term.FgBgColor{Fg: term.White, Bg: term.Red}
				}
			}
			return term.FgBgColor{}
		})
}

func BenchmarkColorLoggerSimple(b *testing.B) {
	benchmarkRunner(b, newColorLogger(b, ioutil.Discard), baseMessage)
}

func BenchmarkColorLoggerContextual(b *testing.B) {
	benchmarkRunner(b, newColorLogger(b, ioutil.Discard), withMessage)
}

func TestColorLoggerConcurrency(t *testing.T) {
	testConcurrency(t, newColorLogger(t, ioutil.Discard))
}

// copied from log/benchmark_test.go
func benchmarkRunner(b *testing.B, logger log.Logger, f func(log.Logger)) {
	lc := log.NewContext(logger).With("common_key", "common_value")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f(lc)
	}
}

var (
	baseMessage = func(logger log.Logger) { logger.Log("foo_key", "foo_value") }
	withMessage = func(logger log.Logger) { log.NewContext(logger).With("a", "b").Log("c", "d") }
)

// copied from log/concurrency_test.go
func testConcurrency(t *testing.T, logger log.Logger) {
	for _, n := range []int{10, 100, 500} {
		wg := sync.WaitGroup{}
		wg.Add(n)
		for i := 0; i < n; i++ {
			go func() { spam(logger); wg.Done() }()
		}
		wg.Wait()
	}
}

func spam(logger log.Logger) {
	for i := 0; i < 100; i++ {
		logger.Log("key", strconv.FormatInt(int64(i), 10))
	}
}
