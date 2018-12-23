package awslambda

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/go-kit/kit/endpoint"
)

func TestServeHTTPLambdaHappyPath(t *testing.T) {
	svc := serviceTest01{}

	helloHandler := NewServer(
		makeTest01HelloEndpoint(svc),
		decodeHelloRequest,
		encodeResponse,
	)

	ctx := context.Background()
	resp, err := helloHandler.ServeHTTPLambda(ctx, events.APIGatewayProxyRequest{
		Body: `{"name":"john doe"}`,
	})

	if err != nil {
		t.Fatalf("\nshould have no error, but got: %+v", err)
	}

	response := helloResponse{}
	err = json.Unmarshal([]byte(resp.Body), &response)
	if err != nil {
		t.Fatalf("\nshould have no error, but got: %+v", err)
	}

	expectedGreeting := "hello john doe"
	if response.Greeting != expectedGreeting {
		t.Fatalf(
			"\nexpect: %s\nactual: %s", expectedGreeting, response.Greeting)
	}
}

func decodeHelloRequest(
	_ context.Context, req events.APIGatewayProxyRequest,
) (interface{}, error) {
	request := helloRequest{}
	err := json.Unmarshal([]byte(req.Body), &request)
	return request, err
}

func encodeResponse(
	_ context.Context, response interface{}, resp events.APIGatewayProxyResponse,
) (events.APIGatewayProxyResponse, error) {
	respByte, err := json.Marshal(response)
	if err != nil {
		return resp, err
	}

	resp.Body = string(respByte)
	resp.StatusCode = 200
	return resp, nil
}

type helloRequest struct {
	Name string `json:"name"`
}

type helloResponse struct {
	Greeting string `json:"greeting"`
}

func makeTest01HelloEndpoint(svc serviceTest01) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(helloRequest)
		greeting := svc.hello(req.Name)
		return helloResponse{greeting}, nil
	}
}

type serviceTest01 struct{}

func (ts *serviceTest01) hello(name string) string {
	return fmt.Sprintf("hello %s", name)
}
