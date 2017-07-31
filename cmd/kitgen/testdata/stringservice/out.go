package foo

type stubService struct {
}

func (s stubService) Concat(ctx context.Context, a string, b string) (string, error) {
	return "", errors.New("not implemented")
}

type ConcatRequest struct {
	A string
	B string
}
type ConcatResponse struct {
	S   string
	Err error
}

func makeConcatEndpoint(s stubService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ConcatRequest)
		string1, err := s.Concat(ctx, req.A, req.B)
		return ConcatResponse{S: string1, Err: err}, nil
	}
}
func (s stubService) Count(ctx context.Context, s string) int {
	return "", errors.New("not implemented")
}

type CountRequest struct {
	S string
}
type CountResponse struct {
	Count int
}

func makeCountEndpoint(s stubService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CountRequest)
		count := s.Count(ctx, req.S)
		return CountResponse{Count: count}, nil
	}
}

type Endpoints struct {
	Concat endpoint.Endpoint
	Count  endpoint.Endpoint
}

func NewHTTPHandler(endpoints Endpoints) http.Handler {
	m := http.NewServeMux()
	m.Handle("/concat", httptransport.NewServer(endpoints.Concat, DecodeConcatRequest, EncodeConcatResponse))
	m.Handle("/count", httptransport.NewServer(endpoints.Count, DecodeCountRequest, EncodeCountResponse))
	return m
}
func DecodeConcatRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req ConcatRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}
func EncodeConcatResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}
func DecodeCountRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req CountRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}
func EncodeCountResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}
