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
	level.AllowInfoAndAbove()
	benchmarkRunner(b, log.NewNopLogger())
}

func BenchmarkNopAllowedLevel(b *testing.B) {
	level.AllowAll()
	benchmarkRunner(b, log.NewNopLogger())
}

func BenchmarkJSONBaseline(b *testing.B) {
	benchmarkRunner(b, log.NewJSONLogger(ioutil.Discard))
}

func BenchmarkJSONDisallowedLevel(b *testing.B) {
	level.AllowInfoAndAbove()
	benchmarkRunner(b, log.NewJSONLogger(ioutil.Discard))
}

func BenchmarkJSONAllowedLevel(b *testing.B) {
	level.AllowAll()
	benchmarkRunner(b, log.NewJSONLogger(ioutil.Discard))
}

func BenchmarkLogfmtBaseline(b *testing.B) {
	benchmarkRunner(b, log.NewLogfmtLogger(ioutil.Discard))
}

func BenchmarkLogfmtDisallowedLevel(b *testing.B) {
	level.AllowInfoAndAbove()
	benchmarkRunner(b, log.NewLogfmtLogger(ioutil.Discard))
}

func BenchmarkLogfmtAllowedLevel(b *testing.B) {
	level.AllowAll()
	benchmarkRunner(b, log.NewLogfmtLogger(ioutil.Discard))
}

func benchmarkRunner(b *testing.B, logger log.Logger) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		level.Debug(logger).Log("foo", "bar")
	}
}
