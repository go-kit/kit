package foo

type stubService struct {
}

func (s stubService) Foo(ctx context.Context, i int, s string) (int, error) {
	return "", errors.New("not implemented")
}

type FooRequest struct {
	I int
	S string
}
type FooResponse struct {
	I   int
	Err error
}

func makeFooEndpoint(s stubService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(FooRequest)
		i, err := s.Foo(ctx, req.I, req.S)
		return FooResponse{I: i, Err: err}, nil
	}
}

type Endpoints struct {
	Foo endpoint.Endpoint
}

func NewHTTPHandler(endpoints Endpoints) http.Handler {
	m := http.NewServeMux()
	m.Handle("/foo", httptransport.NewServer(endpoints.Foo, DecodeFooRequest, EncodeFooResponse))
	return m
}
func DecodeFooRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req FooRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}
func EncodeFooResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}
