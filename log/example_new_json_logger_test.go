package log_test

import (
	"bytes"
	"fmt"

	"github.com/go-kit/kit/log"
)

// JSONLogger Example
func ExampleNewJSONLogger() {
	var buf bytes.Buffer

	logger := log.NewJSONLogger(&buf)
	logger.Log("question", "what is the meaning of life?", "answer", 42)

	fmt.Print(&buf)
	// Output:
	// {"answer":42,"question":"what is the meaning of life?"}
}
