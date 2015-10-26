package http

import (
	"testing"
)

type testBadReauestType struct {
	code int
	msg  string
}

func (err *testBadReauestType) Error() string {
	return err.msg
}

func (err *testBadReauestType) Code() int {
	return err.code
}

func TestBadRequestError(t *testing.T) {
	inner := &testBadReauestType{5432, "bad request custom"}
	err := BadRequestError{inner}

	if inner2, ok := err.Raw().(*testBadReauestType); !ok {
		t.Errorf("want *testBadReauestType have %#v", inner2)
		return
	} else if *inner != *inner2 {
		t.Errorf("want %#v have %#v", inner, inner2)
		return
	}
}
