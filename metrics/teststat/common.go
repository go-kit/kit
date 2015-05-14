// Package teststat contains helper functions for statistical testing of
// metrics implementations.
package teststat

import (
	"math"
	"math/rand"
	"testing"

	"github.com/go-kit/kit/metrics"
)

const population = 1234

// PopulateNormalHistogram populates the Histogram with a normal distribution
// of observations.
func PopulateNormalHistogram(t *testing.T, h metrics.Histogram, seed int64, mean, stdev int64) {
	rand.Seed(seed)
	for i := 0; i < population; i++ {
		sample := int64(rand.NormFloat64()*float64(stdev) + float64(mean))
		h.Observe(sample)
	}
}

// https://en.wikipedia.org/wiki/Normal_distribution#Quantile_function
func normalValueAtQuantile(mean, stdev int64, quantile int) int64 {
	return int64(float64(mean) + float64(stdev)*math.Sqrt2*erfinv(2*(float64(quantile)/100)-1))
}

// https://code.google.com/p/gostat/source/browse/stat/normal.go
func observationsLessThan(mean, stdev int64, x float64, total int) int {
	cdf := ((1.0 / 2.0) * (1 + math.Erf((x-float64(mean))/(float64(stdev)*math.Sqrt2))))
	return int(cdf * float64(total))
}

// https://stackoverflow.com/questions/5971830/need-code-for-inverse-error-function
func erfinv(y float64) float64 {
	if y < -1.0 || y > 1.0 {
		panic("invalid input")
	}

	var (
		a = [4]float64{0.886226899, -1.645349621, 0.914624893, -0.140543331}
		b = [4]float64{-2.118377725, 1.442710462, -0.329097515, 0.012229801}
		c = [4]float64{-1.970840454, -1.624906493, 3.429567803, 1.641345311}
		d = [2]float64{3.543889200, 1.637067800}
	)

	const y0 = 0.7
	var x, z float64

	if math.Abs(y) == 1.0 {
		x = -y * math.Log(0.0)
	} else if y < -y0 {
		z = math.Sqrt(-math.Log((1.0 + y) / 2.0))
		x = -(((c[3]*z+c[2])*z+c[1])*z + c[0]) / ((d[1]*z+d[0])*z + 1.0)
	} else {
		if y < y0 {
			z = y * y
			x = y * (((a[3]*z+a[2])*z+a[1])*z + a[0]) / ((((b[3]*z+b[3])*z+b[1])*z+b[0])*z + 1.0)
		} else {
			z = math.Sqrt(-math.Log((1.0 - y) / 2.0))
			x = (((c[3]*z+c[2])*z+c[1])*z + c[0]) / ((d[1]*z+d[0])*z + 1.0)
		}
		x = x - (math.Erf(x)-y)/(2.0/math.SqrtPi*math.Exp(-x*x))
		x = x - (math.Erf(x)-y)/(2.0/math.SqrtPi*math.Exp(-x*x))
	}

	return x
}
