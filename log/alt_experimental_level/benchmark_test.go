package level_test

import (
	"io/ioutil"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/alt_experimental_level"
)

func BenchmarkNopBaseline(b *testing.B) {
	benchmarkRunner(b, log.NewNopLogger())
}

func BenchmarkNopDisallowedLevel(b *testing.B) {
	benchmarkRunner(b,
		level.AllowingInfoAndAbove(log.NewNopLogger()))
}

func BenchmarkNopAllowedLevel(b *testing.B) {
	benchmarkRunner(b,
		level.AllowingAll(log.NewNopLogger()))
}

func BenchmarkJSONBaseline(b *testing.B) {
	benchmarkRunner(b, log.NewJSONLogger(ioutil.Discard))
}

func BenchmarkJSONDisallowedLevel(b *testing.B) {
	benchmarkRunner(b,
		level.AllowingInfoAndAbove(log.NewJSONLogger(ioutil.Discard)))
}

func BenchmarkJSONAllowedLevel(b *testing.B) {
	benchmarkRunner(b,
		level.AllowingAll(log.NewJSONLogger(ioutil.Discard)))
}

func BenchmarkLogfmtBaseline(b *testing.B) {
	benchmarkRunner(b, log.NewLogfmtLogger(ioutil.Discard))
}

func BenchmarkLogfmtDisallowedLevel(b *testing.B) {
	benchmarkRunner(b,
		level.AllowingInfoAndAbove(log.NewLogfmtLogger(ioutil.Discard)))
}

func BenchmarkLogfmtAllowedLevel(b *testing.B) {
	benchmarkRunner(b,
		level.AllowingAll(log.NewLogfmtLogger(ioutil.Discard)))
}

func benchmarkRunner(b *testing.B, logger log.Logger) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		level.Debug(logger).Log("foo", "bar")
	}
}

func BenchmarkManyDroppedRecords(b *testing.B) {
	logger := level.AllowingInfoAndAbove(log.NewJSONLogger(ioutil.Discard))
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
