package log_test

import (
	"bytes"
	"sync"
	"testing"

	"github.com/peterbourgon/gokit/log"
)

func TestWith(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := log.NewJSONLogger(buf)
	logger = log.With(logger, "a", 123)
	logger = log.With(logger, "b", "c") // With should stack
	if err := logger.Log("msg", "message"); err != nil {
		t.Fatal(err)
	}
	if want, have := `{"a":123,"b":"c","msg":"message"}`+"\n", buf.String(); want != have {
		t.Errorf("want\n\t%#v, have\n\t%#v", want, have)
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
	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < 10000; j++ {
				l.Log("goroutineIdx", idx)
			}
		}(i)
	}
	wg.Wait()

	for _, count := range counts {
		if count != 10000 {
			t.Fatalf("Wrong number of messages in goroutine buckets: %+v", counts)
		}
	}
}
