package log_test

import (
	"bytes"
	"sync"
	"testing"

	"github.com/peterbourgon/gokit/log"
)

func TestWith(t *testing.T) {
	buf := &bytes.Buffer{}
	kvs := []interface{}{"a", 123}
	logger := log.NewJSONLogger(buf)
	logger = log.With(logger, kvs...)
	kvs[1] = 0                          // With should copy its key values
	logger = log.With(logger, "b", "c") // With should stack
	if err := logger.Log("msg", "message"); err != nil {
		t.Fatal(err)
	}
	if want, have := `{"a":123,"b":"c","msg":"message"}`+"\n", buf.String(); want != have {
		t.Errorf("\nwant: %s\nhave: %s", want, have)
	}
}

func TestWither(t *testing.T) {
	logger := &mylogger{}
	log.With(logger, "a", "b").Log("c", "d")
	if want, have := 1, logger.withs; want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

type mylogger struct{ withs int }

func (l *mylogger) Log(keyvals ...interface{}) error { return nil }

func (l *mylogger) With(keyvals ...interface{}) log.Logger { l.withs++; return l }

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
	l := log.With(logger, make([]interface{}, 0, 2)...)

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
	logger := log.NewDiscardLogger()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Log("k", "v")
	}
}

func BenchmarkOneWith(b *testing.B) {
	logger := log.NewDiscardLogger()
	logger = log.With(logger, "k", "v")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Log("k", "v")
	}
}

func BenchmarkTwoWith(b *testing.B) {
	logger := log.NewDiscardLogger()
	for i := 0; i < 2; i++ {
		logger = log.With(logger, "k", "v")
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Log("k", "v")
	}
}

func BenchmarkTenWith(b *testing.B) {
	logger := log.NewDiscardLogger()
	for i := 0; i < 10; i++ {
		logger = log.With(logger, "k", "v")
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Log("k", "v")
	}
}
