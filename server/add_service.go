package server

import (
	"encoding/json"
	"io"
)

// Add TODO
func Add(a, b int) int { return a + b }

// Add_Request TODO
type Add_Request struct {
	A int `json:"a"`
	B int `json:"b"`
}

// Add_Response TODO
type Add_Response struct {
	V int `json:"v"`
}

// Add_Service TODO
func Add_Service(req Request) (Response, error) {
	addReq, ok := req.(Add_Request)
	if !ok {
		return nil, ErrBadCast
	}

	v := Add(addReq.A, addReq.B)
	return Add_Response{v}, nil
}

// Add_Codec_JSON TODO
type Add_Codec_JSON struct{}

// Decode TODO
func (c *Add_Codec_JSON) Decode(src io.Reader) (Request, error) {
	var req Add_Request
	if err := json.NewDecoder(src).Decode(&req); err != nil {
		return nil, err
	}
	return req, nil
}

// Encode TODO
func (c *Add_Codec_JSON) Encode(dst io.Writer, resp Response) error {
	addResp, ok := resp.(Add_Response)
	if !ok {
		return ErrBadCast
	}
	return json.NewEncoder(dst).Encode(addResp)
}
