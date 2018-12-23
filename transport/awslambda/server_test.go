package awslambda

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/go-kit/kit/endpoint"
)

type key int

const (
	KeyBeforeOne key = iota
	KeyBeforeTwo key = iota
)

func TestServeHTTPLambdaHappyPath(t *testing.T) {
	svc := serviceTest01{}

	helloHandler := NewServer(
		makeTest01HelloEndpoint(svc),
		decodeHelloRequest,
		encodeResponse,
		ServerBefore(func(
			ctx context.Context, req events.APIGatewayProxyRequest,
		) context.Context {
			ctx = context.WithValue(ctx, KeyBeforeOne, "bef1")
			return ctx
		}),
		ServerBefore(func(
			ctx context.Context, req events.APIGatewayProxyRequest,
		) context.Context {
			ctx = context.WithValue(ctx, KeyBeforeTwo, "bef2")
			return ctx
		}),
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

	expectedGreeting := "hello john doe bef1 bef2"
	if response.Greeting != expectedGreeting {
		t.Fatalf(
			"\nexpect: %s\nactual: %s", expectedGreeting, response.Greeting)
	}
}

func decodeHelloRequest(
	ctx context.Context, req events.APIGatewayProxyRequest,
) (interface{}, error) {
	request := helloRequest{}
	err := json.Unmarshal([]byte(req.Body), &request)
	if err != nil {
		return request, err
	}

	valOne, ok := ctx.Value(KeyBeforeOne).(string)
	if !ok {
		return request, fmt.Errorf(
			"Value was not set properly when multiple ServerBefores are used")
	}

	valTwo, ok := ctx.Value(KeyBeforeTwo).(string)
	if !ok {
		return request, fmt.Errorf(
			"Value was not set properly when multiple ServerBefores are used")
	}

	request.Name += " " + valOne + " " + valTwo
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
