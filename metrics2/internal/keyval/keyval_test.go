package keyval

import (
	"fmt"
	"testing"

	metrics "github.com/go-kit/kit/metrics2"
	"github.com/google/go-cmp/cmp"
)

func TestMakeWith(t *testing.T) {
	for _, testcase := range []struct {
		input []string
		want  map[string]string
	}{
		{[]string{}, map[string]string{}},
		{[]string{""}, map[string]string{"": metrics.UnknownValue}},
		{[]string{"a"}, map[string]string{"a": metrics.UnknownValue}},
		{[]string{"ab", "cd"}, map[string]string{"ab": metrics.UnknownValue, "cd": metrics.UnknownValue}},
	} {
		t.Run(fmt.Sprintf("%v", testcase.input), func(t *testing.T) {
			if want, have := testcase.want, MakeWith(testcase.input); !cmp.Equal(want, have) {
				t.Errorf("want %v, have %v", want, have)
			}
		})
	}
}

func TestMerge(t *testing.T) {
	original := map[string]string{"a": "b", "c": "d", "e": "f"}
	originalSize := len(original)
	second := Merge(original, "a", "existing key", "x", "new key")
	if want, have := originalSize, len(second); want != have {
		t.Errorf("Merge returns the wrong cardinality: want %d, have %d", want, have)
	}
	if want, have := originalSize, len(original); want != have {
		t.Errorf("Merge modifies the original map: want %d, have %d", want, have)
	}
}
