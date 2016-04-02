package http

import (
	"fmt"
)

// These are some pre-generated constants that can be used to check against
// for the DomainErrors.
const (
	// DomainNewRequest represents an error at the Request Generation
	// Scope.
	DomainNewRequest = "NewRequest"

	// DomainEncode represent an error that has occurred at the Encode
	// level of the request.
	DomainEncode = "Encode"

	// DomainDo represents an error that has occurred at the Do, or
	// execution phase of the request.
	DomainDo = "Do"

	// DomainDecode represents an error that has occured at the Decode
	// phase of the request.
	DomainDecode = "Decode"
)

// TransportError represents an Error occurred in the Client transport level.
type TransportError struct {
	// Domain represents the domain of the error encountered.
	// Simply, this refers to the phase in which the error was
	// generated
	Domain string

	// Err references the underlying error that caused this error
	// overall.
	Err error
}

// Error implements the error interface
func (e TransportError) Error() string {
	return fmt.Sprintf("%s: %s", e.Domain, e.Err)
}
