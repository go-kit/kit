package log

import "io"

// SpecializedWriter represents an extended io.Writer class that can perform
// contextualized writes.
type SpecializedWriter interface {
	GetSpecializedWriter(keyvals ...interface{}) io.Writer
}

// specializedWriter returns a specialized writer for the specified keys or
// falls back to returning w if no specialized writer is available.
func specializedWriter(w io.Writer, keyvals ...interface{}) io.Writer {
	if ew, ok := w.(SpecializedWriter); ok {
		w = ew.GetSpecializedWriter(keyvals)
	}

	return w
}
