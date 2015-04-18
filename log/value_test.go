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

	time.Sleep(50 * time.Millisecond)
	before := time.Now()
	logger.Log()
	after := time.Now()

	if _, ok := output[1].(time.Time); !ok {
		t.Fatalf("output[1] type: want time.Time, have %T", output[1])
	}
	lt := output[1].(time.Time)
	if before.After(lt) {
		t.Errorf("output[1]: want on or after %v, have %v", before, lt)
	}
	if after.Before(lt) {
		t.Errorf("output[1]: want on or before %v, have %v", after, lt)
	}

	if want, have := "value_test.go:21", fmt.Sprint(output[3]); want != have {
		t.Fatalf("output[3]: want %s, have %s", want, have)
	}
}

func BenchmarkValueBindingTime(b *testing.B) {
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
