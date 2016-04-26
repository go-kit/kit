package http

import (
	"fmt"
)

const (
	// DomainNewRequest is an error during request generation.
	DomainNewRequest = "NewRequest"

	// DomainEncode is an error during request or response encoding.
	DomainEncode = "Encode"

	// DomainDo is an error during the execution phase of the request.
	DomainDo = "Do"

	// DomainDecode is an error during request or response decoding.
	DomainDecode = "Decode"
)

// TransportError is an error that occurred at some phase within the transport.
type TransportError struct {
	// Domain is the phase in which the error was generated.
	Domain string

	// Err is the concrete error.
	Err error
}

// Error implements the error interface.
func (e TransportError) Error() string {
	return fmt.Sprintf("%s: %s", e.Domain, e.Err)
}
