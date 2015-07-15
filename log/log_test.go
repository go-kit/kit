package log_test

import (
	"bytes"
	"sync"
	"testing"

	"github.com/go-kit/kit/log"
)

var discard = log.Logger(log.LoggerFunc(func(...interface{}) error { return nil }))

func TestContext(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := log.NewLogfmtLogger(buf)

	kvs := []interface{}{"a", 123}
	lc := log.NewContext(logger).With(kvs...)
	kvs[1] = 0 // With should copy its key values

	lc = lc.With("b", "c") // With should stack
	if err := lc.Log("msg", "message"); err != nil {
		t.Fatal(err)
	}
	if want, have := "a=123 b=c msg=message\n", buf.String(); want != have {
		t.Errorf("\nwant: %shave: %s", want, have)
	}

	buf.Reset()
	lc = lc.WithPrefix("p", "first")
	if err := lc.Log("msg", "message"); err != nil {
		t.Fatal(err)
	}
	if want, have := "p=first a=123 b=c msg=message\n", buf.String(); want != have {
		t.Errorf("\nwant: %shave: %s", want, have)
	}
}

func TestContextWithPrefix(t *testing.T) {
	buf := &bytes.Buffer{}
	kvs := []interface{}{"a", 123}
	logger := log.NewJSONLogger(buf)
	lc := log.NewContext(logger).With(kvs...)
	kvs[1] = 0             // WithPrefix should copy its key values
	lc = lc.With("b", "c") // WithPrefix should stack
	if err := lc.Log("msg", "message"); err != nil {
		t.Fatal(err)
	}
	if want, have := `{"a":123,"b":"c","msg":"message"}`+"\n", buf.String(); want != have {
		t.Errorf("\nwant: %s\nhave: %s", want, have)
	}
}

// Test that With returns a Logger safe for concurrent use. This test
// validates that the stored logging context does not get corrupted when
// multiple clients concurrently log additional keyvals.
//
// This test must be run with go test -cpu 2 (or more) to achieve its goal.
func TestWithConcurrent(t *testing.T) {
	// Create some buckets to count how many events each goroutine logs.
	const goroutines = 8
	counts := [goroutines]int{}

	// This logger extracts a goroutine id from the last value field and
	// increments the referenced bucket.
	logger := log.LoggerFunc(func(kv ...interface{}) error {
		goroutine := kv[len(kv)-1].(int)
		counts[goroutine]++
		return nil
	})

	// With must be careful about handling slices that can grow without
	// copying the underlying array, so give it a challenge.
	l := log.NewContext(logger).With(make([]interface{}, 0, 2)...)

	// Start logging concurrently. Each goroutine logs its id so the logger
	// can bucket the event counts.
	var wg sync.WaitGroup
	wg.Add(goroutines)
	const n = 10000
	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < n; j++ {
				l.Log("goroutineIdx", idx)
			}
		}(i)
	}
	wg.Wait()

	for bucket, have := range counts {
		if want := n; want != have {
			t.Errorf("bucket %d: want %d, have %d", bucket, want, have) // note Errorf
		}
	}
}

func BenchmarkDiscard(b *testing.B) {
	logger := discard
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Log("k", "v")
	}
}

func BenchmarkOneWith(b *testing.B) {
	logger := discard
	lc := log.NewContext(logger).With("k", "v")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lc.Log("k", "v")
	}
}

func BenchmarkTwoWith(b *testing.B) {
	logger := discard
	lc := log.NewContext(logger).With("k", "v")
	for i := 1; i < 2; i++ {
		lc = lc.With("k", "v")
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lc.Log("k", "v")
	}
}

func BenchmarkTenWith(b *testing.B) {
	logger := discard
	lc := log.NewContext(logger).With("k", "v")
	for i := 1; i < 10; i++ {
		lc = lc.With("k", "v")
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lc.Log("k", "v")
	}
}

func TestSwapLogger(t *testing.T) {
	var logger log.SwapLogger

	// Zero value does not panic or error.
	err := logger.Log("k", "v")
	if got, want := err, error(nil); got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	buf := &bytes.Buffer{}
	json := log.NewJSONLogger(buf)
	logger.Swap(json)

	if err := logger.Log("k", "v"); err != nil {
		t.Error(err)
	}
	if got, want := buf.String(), `{"k":"v"}`+"\n"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	buf.Reset()
	prefix := log.NewLogfmtLogger(buf)
	logger.Swap(prefix)

	if err := logger.Log("k", "v"); err != nil {
		t.Error(err)
	}
	if got, want := buf.String(), "k=v\n"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	buf.Reset()
	logger.Swap(nil)

	if err := logger.Log("k", "v"); err != nil {
		t.Error(err)
	}
	if got, want := buf.String(), ""; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestSwapLoggerConcurrency(t *testing.T) {
	testConcurrency(t, &log.SwapLogger{})
}
