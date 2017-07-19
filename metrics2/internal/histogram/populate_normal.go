package histogram

import "math/rand"

// PopulateNormal makes a series of normal random observations into the
// histogram. The number of observations is determined by count. The
// distribution is determined by mean and stdev. Randomness is controlled by
// seed.
func PopulateNormal(o Observable, count, mean, stdev, seed int) {
	r := rand.New(rand.NewSource(int64(seed)))
	for i := 0; i < count; i++ {
		sample := r.NormFloat64()*float64(stdev) + float64(mean)
		if sample < 0 {
			sample = 0
		}
		o.Observe(sample)
	}
}

// Observable models the write side of a histogram.
type Observable interface {
	Observe(float64)
}
