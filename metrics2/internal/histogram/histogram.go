package histogram

import (
	"fmt"
	"io"
	"strings"

	"gopkg.in/caio/go-tdigest.v1"
)

// Histogram accumulates observations and provides quantile summaries of
// observations. Note that it's implemented with a TDigest, which
type Histogram struct {
	d *tdigest.TDigest
}

// New returns a Histogram with a default compression ratio.
func New() *Histogram {
	return NewCompression(100.0)
}

// NewCompression returns a Histogram with the given compression ratio.
// Compression must be > 1.0.
func NewCompression(compression float64) *Histogram {
	return &Histogram{
		d: tdigest.New(compression),
	}
}

// Observe a value into the histogram.
func (h *Histogram) Observe(value float64) {
	h.d.Add(value, 1)
}

// Quantile returns the given quantile (0..1) of the histogram.
func (h *Histogram) Quantile(q float64) float64 {
	return h.d.Quantile(q)
}

// Bucketize the observations into bucketCount buckets.
func (h *Histogram) Bucketize(bucketCount int) []Bucket {
	var (
		agg      = map[float64]uint32{}
		min, max float64
	)
	h.d.ForEachCentroid(func(mean float64, count uint32) bool {
		if len(agg) == 0 {
			min, max = mean, mean
		}
		if mean < min {
			min = mean
		}
		if mean > max {
			max = mean
		}
		agg[mean] += count
		return true
	})
	min -= 0.0001
	max += 0.0001
	var (
		delta   = max - min
		step    = delta / float64(bucketCount)
		buckets = make([]Bucket, bucketCount)
	)
	for i := 0; i < bucketCount; i++ {
		maxValue := min + (step * float64(i+1))
		buckets[i].Max = maxValue
		for value, count := range agg {
			if value < maxValue {
				buckets[i].Count += count
			}
		}
	}
	return buckets
}

// Bucket combines a max value and a count of observations less than that bucket.
type Bucket struct {
	Max   float64
	Count uint32
}

// Render a simple bar chart to the writer.
func Render(w io.Writer, buckets []Bucket, width int) {
	if width < 10 {
		width = 10
	}
	var prev, max uint32
	for _, b := range buckets {
		count := b.Count - prev
		if count > max {
			max = count
		}
		prev = b.Count
	}
	prev = 0
	for _, b := range buckets {
		var (
			count   = b.Count - prev
			percent = float64(count) / float64(max)
			bars    = int(percent * float64(width))
			str     = strings.Repeat("#", bars)
		)
		fmt.Fprintf(w, "%.4f: %s\n", b.Max, str)
		prev = b.Count
	}
}
