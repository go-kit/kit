package foo

type STUBSTRUCT struct{}

func (f stubFooService) Bar(ctx context.Context, i int, s string) (string, error) {
	return "", errors.New("not implemented")
}

type BarRequest struct {
	I int
	S string
}

type BarResponse struct {
	S   string
	Err error
}

func makeBarEndpoint(s stubFooService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(barrequest)
		s, err := s.bar(ctx, req.i, req.s)
		return barresponse{s: s, err: err}, nil
	}
}

type Endpoints struct {
	Bar endpoint.Endpoint
}

// Each transport binding should be opt-in with a flag to kitgen.
// Here's a basic sketch of what HTTP may look like.
// n.b. comments should encourage users to edit the generated code.

func NewHTTPHandler(endpoints Endpoints) http.Handler {
	m := http.NewServeMux()
	m.Handle("/bar", httptransport.NewServer(
		endpoints.Bar,
		DecodeBarRequest,
		EncodeBarResponse,
	))
	return m
}

func DecodeBarRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req BarRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

func EncodeBarResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}
