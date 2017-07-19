package keyval

import "testing"

func TestAppend(t *testing.T) {
	original := map[string]string{"a": "b", "c": "d"}
	originalSize := len(original)
	second := Append(original, "foo", "bar", "baz", "quux")
	if want, have := originalSize+2 /* pairs */, len(second); want != have {
		t.Errorf("Append returns the wrong cardinality: want %d, have %d", want, have)
	}
	if want, have := originalSize, len(original); want != have {
		t.Errorf("Append modifies the original map: want %d, have %d", want, have)
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
