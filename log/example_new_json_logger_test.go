package log_test

import (
	"bytes"
	"fmt"

	"github.com/go-kit/kit/log"
)

// JSONLogger Example
func ExampleNewJSONLogger() {
	log.NewJSONLogger(os.Stderr).Log("meaning of life", 42)
	// Output:
	// {"meaning_of_life":42}
}
