package metrics_test

import (
	"bytes"
	"testing"

	"math"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/expvar"
	"github.com/go-kit/kit/metrics/teststat"
)

func TestPrintDistribution(t *testing.T) {
	var (
		quantiles = []int{50, 90, 95, 99}
		h         = expvar.NewHistogram("test_print_distribution", 0, 100, 3, quantiles...)
		seed      = int64(555)
		mean      = int64(5)
		stdev     = int64(1)
	)
	teststat.PopulateNormalHistogram(t, h, seed, mean, stdev)

	var buf bytes.Buffer
	metrics.PrintDistribution(&buf, h)
	t.Logf("\n%s\n", buf.String())

	// Count the number of bar chart characters.
	// We should have ca. 100 in any distribution with a small-enough stdev.

	var n int
	for _, r := range buf.String() {
		if r == '#' {
			n++
		}
	}
	if want, have, tol := 100, n, 5; int(math.Abs(float64(want-have))) > tol {
		t.Errorf("want %d, have %d (tolerance %d)", want, have, tol)
	}
}
