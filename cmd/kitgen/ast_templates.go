// This file was automatically generated based on the contents of *.tmpl
// If you need to update this file, change the contents of those files
// (or add new ones) and run 'go generate'

package main

import "golang.org/x/tools/godoc/vfs/mapfs"

var ASTTemplates = mapfs.New(map[string]string{
	`full.go`: "package foo\n\ntype STUBSTRUCT struct{}\n\nfunc (f STUBSTRUCT) Bar(PARAM aType) (string, error) {\n	return \"\", errors.New(\"not implemented\")\n}\n\ntype BarRequest struct {\n	I int\n	S string\n}\n\ntype BarResponse struct {\n	S   string\n	Err error\n}\n\nfunc makeBarEndpoint(s STUBSTRUCT) endpoint.Endpoint {\n	return func(ctx context.Context, request interface{}) (interface{}, error) {\n		req := request.(barrequest)\n		s, err := s.bar(ctx, req.i, req.s)\n		return barresponse{s: s, err: err}, nil\n	}\n}\n\ntype Endpoints struct {\n	Bar endpoint.Endpoint\n}\n\n// Each transport binding should be opt-in with a flag to kitgen.\n// Here's a basic sketch of what HTTP may look like.\n// n.b. comments should encourage users to edit the generated code.\n\nfunc NewHTTPHandler(endpoints Endpoints) http.Handler {\n	m := http.NewServeMux()\n	m.Handle(\"/bar\", httptransport.NewServer(\n		endpoints.Bar,\n		DecodeBarRequest,\n		EncodeBarResponse,\n	))\n	return m\n}\n\nfunc DecodeBarRequest(_ context.Context, r *http.Request) (interface{}, error) {\n	var req BarRequest\n	err := json.NewDecoder(r.Body).Decode(&req)\n	return req, err\n}\n\nfunc EncodeBarResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {\n	w.Header().Set(\"Content-Type\", \"application/json; charset=utf-8\")\n	return json.NewEncoder(w).Encode(response)\n}\n",
})
