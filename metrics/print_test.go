package metrics_test

import (
	"os"
	"testing"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/expvar"
	"github.com/go-kit/kit/metrics/teststat"
)

func TestPrintDistribution(t *testing.T) {
	var (
		name      = "foobar"
		quantiles = []int{50, 90, 95, 99}
		h         = expvar.NewHistogram("test_print_distribution", 1, 10, 3, quantiles...)
		seed      = int64(555)
		mean      = int64(5)
		stdev     = int64(1)
	)
	teststat.PopulateNormalHistogram(t, h, seed, mean, stdev)
	metrics.PrintDistribution(os.Stdout, name, h.Distribution())
}
