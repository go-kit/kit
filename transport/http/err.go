package http

import (
	"fmt"
)

// Err represents an Error occurred in the Client transport level.
type Err struct {
	// Domain represents the domain of the error encountered.
	// Simply, this refers to the phase in which the error was
	// generated
	Domain string

	// Err references the underlying error that caused this error
	// overall.
	Err error
}

// Error implements the error interface
func (e Err) Error() string {
	return fmt.Sprintf("%s: %s", e.Domain, e.Err)
}
