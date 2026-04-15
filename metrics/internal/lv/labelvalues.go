package lv

// LabelValues is a type alias that provides validation on its With method.
// Metrics may include it as a member to help them satisfy With semantics and
// save some code duplication.
type LabelValues []string

// With validates the input, and returns a new aggregate labelValues.
func (lvs LabelValues) With(labelValues ...string) LabelValues {
	if len(labelValues)%2 != 0 {
		labelValues = append(labelValues, "unknown")
	}
	// Create a new slice to avoid aliasing the underlying array.
	// Without this, sibling With calls (e.g., c1.With("a","1") and
	// c1.With("b","2")) can overwrite each other's label values.
	result := make(LabelValues, len(lvs), len(lvs)+len(labelValues))
	copy(result, lvs)
	return append(result, labelValues...)
}
