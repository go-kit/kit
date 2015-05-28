package log_test

import (
	"bytes"
	"fmt"

	"github.com/go-kit/kit/log"
)

// PrefixLogger Example
func ExampleNewPrefixLogger() {
	log.NewPrefixLogger(os.Stderr).Log("question", "What is the meaning of life?", "answer", 42)
	// Output:
	// question=what is the meaning of life? answer=42
}
