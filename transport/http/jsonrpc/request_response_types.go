package jsonrpc

import "encoding/json"

// Request defines a JSON RPC request from the spec
// http://www.jsonrpc.org/specification#request_object
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	ID      interface{}     `json:"id"`
}

// Response defines a JSON RPC response from the spec
// http://www.jsonrpc.org/specification#response_object
type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   Error       `json:"error,omitemty"`
}

const (
	// Version defines the version of the JSON RPC implementation
	Version string = "2.0"

	// ContentType defines the content type to be served.
	ContentType string = "application/json; charset=utf-8"
)
