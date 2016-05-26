package teststat

import (
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/codahale/hdrhistogram"
)

// AssertCirconusNormalHistogram ensures the Circonus Histogram data captured in
// the result slice abides a normal distribution.
func AssertCirconusNormalHistogram(t *testing.T, mean, stdev, min, max int64, result []string) {
	if len(result) <= 0 {
		t.Fatal("no results")
	}

	// Circonus just dumps the raw counts. We need to do our own statistical analysis.
	h := hdrhistogram.New(min, max, 3)

	for _, s := range result {
		// "H[1.23e04]=123"
		toks := strings.Split(s, "=")
		if len(toks) != 2 {
			t.Fatalf("bad H value: %q", s)
		}

		var bucket string
		bucket = toks[0]
		bucket = bucket[2 : len(bucket)-1] // "H[1.23e04]" -> "1.23e04"
		f, err := strconv.ParseFloat(bucket, 64)
		if err != nil {
			t.Fatalf("error parsing H value: %q: %v", s, err)
		}

		count, err := strconv.ParseFloat(toks[1], 64)
		if err != nil {
			t.Fatalf("error parsing H count: %q: %v", s, err)
		}

		h.RecordValues(int64(f), int64(count))
	}

	// Apparently Circonus buckets observations by dropping a sigfig, so we have
	// very coarse tolerance.
	var tolerance int64 = 30
	for _, quantile := range []int{50, 90, 99} {
		want := normalValueAtQuantile(mean, stdev, quantile)
		have := h.ValueAtQuantile(float64(quantile))
		if int64(math.Abs(float64(want)-float64(have))) > tolerance {
			t.Errorf("quantile %d: want %d, have %d", quantile, want, have)
		}
	}
}
