package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// FieldEncoder describes a way to encode fields.
type FieldEncoder func(io.Writer, Field) error

// KeyEqualsValue is a field encoder that yields "key=value".
func KeyEqualsValue(w io.Writer, f Field) error {
	_, err := fmt.Fprintf(w, "%s=%v", f.Key, f.Value)
	return err
}

// ValueOnly is a field encoder that yields "value".
func ValueOnly(w io.Writer, f Field) error {
	_, err := fmt.Fprintf(w, "%v", f.Value)
	return err
}

// JSON is a field encoder that yields '{"key":"value"}' without a trailing
// newline.
func JSON(w io.Writer, f Field) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(map[string]interface{}{f.Key: f.Value}); err != nil {
		return err
	}
	buf.Truncate(buf.Len() - 1) // remove trailing newline
	_, err := buf.WriteTo(w)
	return err
}

// EncodeMany uses the field encoder to serialize the fields to the writer,
// putting a space character between each serialized field.
func EncodeMany(w io.Writer, e FieldEncoder, fields []Field) error {
	for _, f := range fields {
		if err := e(w, f); err != nil {
			return err
		}
		if _, err := fmt.Fprint(w, " "); err != nil {
			return err
		}
	}
	return nil
}
