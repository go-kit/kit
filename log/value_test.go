package log_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/peterbourgon/gokit/log"
)

func TestValueBinding(t *testing.T) {
	var output []interface{}

	logger := log.Logger(log.LoggerFunc(func(keyvals ...interface{}) error {
		output = keyvals
		return nil
	}))
	logger = log.With(logger, "ts", log.Timestamp, "caller", log.Caller)

	before := time.Now()
	logger.Log("foo", "bar")
	after := time.Now()

	timestamp, ok := output[1].(time.Time)
	if !ok {
		t.Fatalf("want time.Time, have %T", output[1])
	}
	if before.After(timestamp) {
		t.Errorf("before %v is after timestamp %v", before, timestamp)
	}
	if after.Before(timestamp) {
		t.Errorf("after %v is before timestamp %v", after, timestamp)
	}

	if want, have := "value_test.go:21", fmt.Sprint(output[3]); want != have {
		t.Fatalf("output[3]: want %s, have %s", want, have)
	}
}

func BenchmarkValueBindingTimestamp(b *testing.B) {
	logger := log.NewDiscardLogger()
	logger = log.With(logger, "ts", log.Timestamp)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Log("k", "v")
	}
}

func BenchmarkValueBindingCaller(b *testing.B) {
	logger := log.NewDiscardLogger()
	logger = log.With(logger, "caller", log.Caller)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Log("k", "v")
	}
}
