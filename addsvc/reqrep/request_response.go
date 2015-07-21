package reqrep

// The concrete request and response types are defined for each method our
// service implements. Request types should be annotated sufficiently for all
// transports we intend to use.

// AddRequest is a request for the add method.
type AddRequest struct {
	A int64 `json:"a"`
	B int64 `json:"b"`
}

// AddResponse is a response to the add method.
type AddResponse struct {
	V int64 `json:"v"`
}
