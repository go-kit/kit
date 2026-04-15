package lv

import (
	"strings"
	"testing"
)

func TestWith(t *testing.T) {
	var a LabelValues
	b := a.With("a", "1")
	c := a.With("b", "2", "c", "3")

	if want, have := "", strings.Join(a, ""); want != have {
		t.Errorf("With appears to mutate the original LabelValues: want %q, have %q", want, have)
	}
	if want, have := "a1", strings.Join(b, ""); want != have {
		t.Errorf("With does not appear to return the right thing: want %q, have %q", want, have)
	}
	if want, have := "b2c3", strings.Join(c, ""); want != have {
		t.Errorf("With does not appear to return the right thing: want %q, have %q", want, have)
	}
}

func TestWithNoAliasing(t *testing.T) {
	base := LabelValues{"a", "1", "b", "2"}
	c1 := base.With("c", "3", "d", "4")
	c2 := base.With("e", "5")
	
	// c2 should not have overwritten c1's values
	if c1[4] != "c" || c1[5] != "3" {
		t.Errorf("c1 was corrupted by sibling With call: got %v", c1)
	}
	if c2[4] != "e" || c2[5] != "5" {
		t.Errorf("c2 has wrong values: got %v", c2)
	}
}
