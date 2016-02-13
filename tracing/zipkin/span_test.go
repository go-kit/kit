package zipkin_test

import (
	"testing"

	"github.com/go-kit/kit/tracing/zipkin"
)

func TestAnnotateBinaryEncodesKeyValueAsBytes(t *testing.T) {
	key := "awesome-bytes-test"
	value := []byte("this is neat")

	span := &zipkin.Span{}
	span.AnnotateBinary(key, value)

	encodedSpan := span.Encode()
	annotations := encodedSpan.GetBinaryAnnotations()

	if annotations[0].Key != key {
		t.Errorf("Error: expected %s got %s", key, annotations[0].Key)
	}

	if string(annotations[0].Value) != string(value) {
		t.Errorf("Error: expected %s got %s", string(value), string(annotations[0].Value))
	}
}

func TestAnnotateStringEncodesKeyValueAsBytes(t *testing.T) {
	key := "awesome-string-test"
	value := "this is neat"

	span := &zipkin.Span{}
	span.AnnotateString(key, value)

	encodedSpan := span.Encode()
	annotations := encodedSpan.GetBinaryAnnotations()

	if annotations[0].Key != key {
		t.Errorf("Error: expected %s got %s", key, annotations[0].Key)
	}

	if string(annotations[0].Value) != value {
		t.Errorf("Error: expected %s got %s", value, string(annotations[0].Value))
	}
}
