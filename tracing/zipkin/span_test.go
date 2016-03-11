package zipkin_test

import (
	"bytes"
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

	if len(annotations) == 0 {
		t.Error("want non-zero length slice, have empty slice")
	}

	if want, have := key, annotations[0].Key; want != have {
		t.Errorf("want %q, got %q", want, have)
	}

	if want, have := value, annotations[0].Value; bytes.Compare(want, have) != 0 {
		t.Errorf("want %s, got %s", want, have)
	}
}

func TestAnnotateStringEncodesKeyValueAsBytes(t *testing.T) {
	key := "awesome-string-test"
	value := "this is neat"

	span := &zipkin.Span{}
	span.AnnotateString(key, value)

	encodedSpan := span.Encode()
	annotations := encodedSpan.GetBinaryAnnotations()

	if len(annotations) == 0 {
		t.Error("want non-zero length slice, have empty slice")
	}

	if want, have := key, annotations[0].Key; want != have {
		t.Errorf("want %q, got %q", want, have)
	}

	if want, have := value, annotations[0].Value; bytes.Compare([]byte(want), have) != 0 {
		t.Errorf("want %s, got %s", want, have)
	}
}
