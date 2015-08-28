package server

// SumRequest is the business domain type for a Sum method request.
type SumRequest struct {
	A int `json:"a"`
	B int `json:"b"`
}

// SumResponse is the business domain type for a Sum method response.
type SumResponse struct {
	V int `json:"v"`
}

// ConcatRequest is the business domain type for a Concat method request.
type ConcatRequest struct {
	A string `json:"a"`
	B string `json:"b"`
}

// ConcatResponse is the business domain type for a Concat method response.
type ConcatResponse struct {
	V string `json:"v"`
}
