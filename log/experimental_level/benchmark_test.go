package level_test

import (
	"io/ioutil"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/experimental_level"
)

func BenchmarkNopBaseline(b *testing.B) {
	benchmarkRunner(b, log.NewNopLogger())
}

func BenchmarkNopDisallowedLevel(b *testing.B) {
	benchmarkRunner(b, level.New(log.NewNopLogger(),
		level.Allowed(level.AllowInfoAndAbove())))
}

func BenchmarkNopAllowedLevel(b *testing.B) {
	benchmarkRunner(b, level.New(log.NewNopLogger(),
		level.Allowed(level.AllowAll())))
}

func BenchmarkJSONBaseline(b *testing.B) {
	benchmarkRunner(b, log.NewJSONLogger(ioutil.Discard))
}

func BenchmarkJSONDisallowedLevel(b *testing.B) {
	benchmarkRunner(b, level.New(log.NewJSONLogger(ioutil.Discard),
		level.Allowed(level.AllowInfoAndAbove())))
}

func BenchmarkJSONAllowedLevel(b *testing.B) {
	benchmarkRunner(b, level.New(log.NewJSONLogger(ioutil.Discard),
		level.Allowed(level.AllowAll())))
}

func BenchmarkLogfmtBaseline(b *testing.B) {
	benchmarkRunner(b, log.NewLogfmtLogger(ioutil.Discard))
}

func BenchmarkLogfmtDisallowedLevel(b *testing.B) {
	benchmarkRunner(b, level.New(log.NewLogfmtLogger(ioutil.Discard),
		level.Allowed(level.AllowInfoAndAbove())))
}

func BenchmarkLogfmtAllowedLevel(b *testing.B) {
	benchmarkRunner(b, level.New(log.NewLogfmtLogger(ioutil.Discard),
		level.Allowed(level.AllowAll())))
}

func benchmarkRunner(b *testing.B, logger log.Logger) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		level.Debug(logger).Log("foo", "bar")
	}
}

func BenchmarkManyDroppedRecords(b *testing.B) {
	logger := level.New(log.NewJSONLogger(ioutil.Discard), level.Config{
		Allowed: level.AllowInfoAndAbove(),
	})
	b.ResetTimer()
	b.ReportAllocs()
	debug := level.Debug(logger)
	info := level.Info(logger)
	for i := 0; i < b.N; i++ {
		debug.Log("foo", "1")
		// Only this one will be retained.
		info.Log("baz", "quux")
		debug.Log("foo", "2")
		debug.Log("foo", "3")
		debug.Log("foo", "4")
	}
}
