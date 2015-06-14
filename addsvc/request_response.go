package main

// The concrete request and response types are defined for each method our
// service implements. Request types should be annotated sufficiently for all
// transports we intend to use.

type addRequest struct {
	A int64 `json:"a"`
	B int64 `json:"b"`
}

type addResponse struct {
	V int64 `json:"v"`
}
