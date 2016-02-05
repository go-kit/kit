package metrics

import (
	"fmt"
	"io"
	"text/tabwriter"
)

const (
	bs  = "####################################################################################################"
	bsz = float64(len(bs))
)

// PrintDistribution writes a human-readable graph of the distribution to the
// passed writer.
func PrintDistribution(w io.Writer, h Histogram) {
	buckets, quantiles := h.Distribution()

	fmt.Fprintf(w, "name: %v\n", h.Name())
	fmt.Fprintf(w, "quantiles: %v\n", quantiles)

	var total float64
	for _, bucket := range buckets {
		total += float64(bucket.Count)
	}

	tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
	fmt.Fprintf(tw, "From\tTo\tCount\tProb\tBar\n")

	axis := "|"
	for _, bucket := range buckets {
		if bucket.Count > 0 {
			p := float64(bucket.Count) / total
			fmt.Fprintf(tw, "%d\t%d\t%d\t%.4f\t%s%s\n", bucket.From, bucket.To, bucket.Count, p, axis, bs[:int(p*bsz)])
			axis = "|"
		} else {
			axis = ":" // show that some bars were skipped
		}
	}

	tw.Flush()
}
