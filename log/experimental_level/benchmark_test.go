package level_test

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/experimental_level"
)

func BenchmarkNopBaseline(b *testing.B) {
	singleRecordBenchmarkRunner(b, log.NewNopLogger())
}

func BenchmarkNopDisallowedLevel(b *testing.B) {
	singleRecordBenchmarkRunner(b, level.New(log.NewNopLogger(),
		level.Allowed(level.AllowInfoAndAbove())))
}

func BenchmarkNopAllowedLevel(b *testing.B) {
	singleRecordBenchmarkRunner(b, level.New(log.NewNopLogger(),
		level.Allowed(level.AllowAll())))
}

func BenchmarkJSONBaseline(b *testing.B) {
	singleRecordBenchmarkRunner(b, log.NewJSONLogger(ioutil.Discard))
}

func BenchmarkJSONDisallowedLevel(b *testing.B) {
	singleRecordBenchmarkRunner(b, level.New(log.NewJSONLogger(ioutil.Discard),
		level.Allowed(level.AllowInfoAndAbove())))
}

func BenchmarkJSONAllowedLevel(b *testing.B) {
	singleRecordBenchmarkRunner(b, level.New(log.NewJSONLogger(ioutil.Discard),
		level.Allowed(level.AllowAll())))
}

func BenchmarkLogfmtBaseline(b *testing.B) {
	singleRecordBenchmarkRunner(b, log.NewLogfmtLogger(ioutil.Discard))
}

func BenchmarkLogfmtDisallowedLevel(b *testing.B) {
	singleRecordBenchmarkRunner(b, level.New(log.NewLogfmtLogger(ioutil.Discard),
		level.Allowed(level.AllowInfoAndAbove())))
}

func BenchmarkLogfmtAllowedLevel(b *testing.B) {
	singleRecordBenchmarkRunner(b, level.New(log.NewLogfmtLogger(ioutil.Discard),
		level.Allowed(level.AllowAll())))
}

func singleRecordBenchmarkRunner(b *testing.B, logger log.Logger) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		level.Debug(logger).Log("foo", "bar")
	}
}

func BenchmarkDroppedRecords(b *testing.B) {
	logger := log.NewNopLogger()
	logger = log.NewContext(logger).With("ts", log.DefaultTimestamp, "caller", log.DefaultCaller)
	for _, dropped := range []uint{1, 3, 9, 99, 999} {
		b.Run(fmt.Sprintf("%d-of-%d", dropped, dropped+1), func(b *testing.B) {
			manyRecordBenchmarkRunner(b, logger, dropped)
		})
	}
}

func manyRecordBenchmarkRunner(b *testing.B, logger log.Logger, droppedRecords uint) {
	logger = level.New(logger, level.Allowed(level.AllowInfoAndAbove()))
	debug := level.Debug(logger)
	info := level.Info(logger)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Only this one will be retained.
		info.Log("foo", "bar")
		for dropped := droppedRecords; dropped != 0; dropped-- {
			debug.Log("baz", "quux")
		}
	}
}
