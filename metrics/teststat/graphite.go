package teststat

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"testing"
)

// AssertGraphiteNormalHistogram ensures the expvar Histogram referenced by
// metricName abides a normal distribution.
func AssertGraphiteNormalHistogram(t *testing.T, prefix, metricName string, mean, stdev int64, quantiles []int, gPayload string) {
	const tolerance int = 2

	// check for hdr histo data
	wants := map[string]int64{"count": 1234, "min": 15, "max": 83, "std-dev": stdev, "mean": mean}
	for key, want := range wants {
		re := regexp.MustCompile(fmt.Sprintf("%s.%s.%s (\\d*)", prefix, metricName, key))
		if res := re.FindAllStringSubmatch(gPayload, 1); res != nil {
			if len(res[0]) == 1 {
				t.Errorf("bad regex found, please check the test scenario")
				continue
			}

			have, err := strconv.ParseInt(res[0][1], 10, 64)
			if err != nil {
				t.Fatal(err)
			}

			if int(math.Abs(float64(want-have))) > tolerance {
				t.Errorf("key %s: want %d, have %d", key, want, have)
			}
		} else {
			t.Error("did not find metrics log for", key, "in \n", gPayload)
		}
	}

	// check for quantile gauges
	for _, quantile := range quantiles {
		want := normalValueAtQuantile(mean, stdev, quantile)

		re := regexp.MustCompile(fmt.Sprintf("%s.%s_p%02d (\\d*\\.\\d*)", prefix, metricName, quantile))
		if res := re.FindAllStringSubmatch(gPayload, 1); res != nil {
			if len(res[0]) == 1 {
				t.Errorf("bad regex found, please check the test scenario")
				continue
			}
			have, err := strconv.ParseFloat(res[0][1], 64)
			if err != nil {
				t.Fatal(err)
			}
			if int(math.Abs(float64(want)-have)) > tolerance {
				t.Errorf("quantile %d: want %.2f, have %.2f", quantile, want, have)
			}
		} else {
			t.Errorf("did not find metrics log for %d", quantile)
		}

	}
}
