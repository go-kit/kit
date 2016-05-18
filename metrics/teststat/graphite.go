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
	// check for hdr histo data
	wants := map[string]int64{"count": 1234, "min": 15, "max": 83}
	for key, want := range wants {
		re := regexp.MustCompile(fmt.Sprintf("%s%s.%s (\\d*)", prefix, metricName, key))
		res := re.FindAllStringSubmatch(gPayload, 1)
		if res == nil {
			t.Error("did not find metrics log for", key, "in \n", gPayload)
			continue
		}

		if len(res[0]) == 1 {
			t.Fatalf("%q: bad regex, please check the test scenario", key)
		}

		have, err := strconv.ParseInt(res[0][1], 10, 64)
		if err != nil {
			t.Fatal(err)
		}

		if want != have {
			t.Errorf("key %s: want %d, have %d", key, want, have)
		}
	}

	const tolerance int = 2
	wants = map[string]int64{".std-dev": stdev, ".mean": mean}
	for _, quantile := range quantiles {
		wants[fmt.Sprintf("_p%02d", quantile)] = normalValueAtQuantile(mean, stdev, quantile)
	}
	// check for quantile gauges
	for key, want := range wants {
		re := regexp.MustCompile(fmt.Sprintf("%s%s%s (\\d*\\.\\d*)", prefix, metricName, key))
		res := re.FindAllStringSubmatch(gPayload, 1)
		if res == nil {
			t.Errorf("did not find metrics log for %s", key)
			continue
		}

		if len(res[0]) == 1 {
			t.Fatalf("%q: bad regex found, please check the test scenario", key)
		}
		have, err := strconv.ParseFloat(res[0][1], 64)
		if err != nil {
			t.Fatal(err)
		}
		if int(math.Abs(float64(want)-have)) > tolerance {
			t.Errorf("key %s: want %.2f, have %.2f", key, want, have)
		}
	}
}
