package awslambda

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

type key int

const (
	KeyBeforeOne key = iota
	KeyBeforeTwo key = iota
	KeyAfterOne  key = iota
	KeyEncMode   key = iota
)

func TestServeHTTPLambdaHappyPath(t *testing.T) {
	svc := serviceTest01{}

	helloHandler := NewServer(
		makeTest01HelloEndpoint(svc),
		decodeHelloRequest,
		encodeResponse,
		ServerErrorLogger(log.NewNopLogger()),
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
		ServerAfter(func(
			ctx context.Context, resp events.APIGatewayProxyResponse,
		) context.Context {
			ctx = context.WithValue(ctx, KeyAfterOne, "af1")
			return ctx
		}),
		ServerAfter(func(
			ctx context.Context, resp events.APIGatewayProxyResponse,
		) context.Context {
			if _, ok := ctx.Value(KeyAfterOne).(string); !ok {
				t.Fatalf("\nValue was not set properly during multi ServerAfter")
			}
			return ctx
		}),
		ServerFinalizer(func(
			_ context.Context, resp events.APIGatewayProxyResponse, _ error,
		) {
			response := helloResponse{}
			err := json.Unmarshal([]byte(resp.Body), &response)
			if err != nil {
				t.Fatalf("\nshould have no error, but got: %+v", err)
			}

			expectedGreeting := "hello john doe bef1 bef2"
			if response.Greeting != expectedGreeting {
				t.Fatalf(
					"\nexpect: %s\nactual: %s", expectedGreeting, response.Greeting)
			}
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

func TestServeHTTPLambdaFailDecode(t *testing.T) {
	svc := serviceTest01{}

	helloHandler := NewServer(
		makeTest01HelloEndpoint(svc),
		decodeHelloRequest,
		encodeResponse,
		ServerErrorEncoder(func(
			ctx context.Context, err error, resp events.APIGatewayProxyResponse,
		) (events.APIGatewayProxyResponse, error) {
			resp.Body = `{"error":"yes"}`
			resp.StatusCode = 500
			return resp, err
		}),
	)

	ctx := context.Background()
	resp, err := helloHandler.ServeHTTPLambda(ctx, events.APIGatewayProxyRequest{
		Body: `{"name":"john doe"}`,
	})

	if err == nil {
		t.Fatalf("\nshould have error, but got: %+v", err)
	}

	if resp.StatusCode != 500 {
		t.Fatalf("\nexpect status code of 500, instead of %d", resp.StatusCode)
	}
}

func TestServeHTTPLambdaFailEndpoint(t *testing.T) {
	svc := serviceTest01{}

	helloHandler := NewServer(
		makeTest01FailEndpoint(svc),
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
		ServerErrorEncoder(func(
			ctx context.Context, err error, resp events.APIGatewayProxyResponse,
		) (events.APIGatewayProxyResponse, error) {
			resp.Body = `{"error":"yes"}`
			resp.StatusCode = 500
			return resp, err
		}),
	)

	ctx := context.Background()
	resp, err := helloHandler.ServeHTTPLambda(ctx, events.APIGatewayProxyRequest{
		Body: `{"name":"john doe"}`,
	})

	if err == nil {
		t.Fatalf("\nshould have error, but got: %+v", err)
	}

	if resp.StatusCode != 500 {
		t.Fatalf("\nexpect status code of 500, instead of %d", resp.StatusCode)
	}
}

func TestServeHTTPLambdaFailEncode(t *testing.T) {
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
		ServerAfter(func(
			ctx context.Context, resp events.APIGatewayProxyResponse,
		) context.Context {
			ctx = context.WithValue(ctx, KeyEncMode, "fail_encode")
			return ctx
		}),
		ServerErrorEncoder(func(
			ctx context.Context, err error, resp events.APIGatewayProxyResponse,
		) (events.APIGatewayProxyResponse, error) {
			resp.Body = `{"error":"yes"}`
			resp.StatusCode = 500
			return resp, err
		}),
	)

	ctx := context.Background()
	resp, err := helloHandler.ServeHTTPLambda(ctx, events.APIGatewayProxyRequest{
		Body: `{"name":"john doe"}`,
	})

	if err == nil {
		t.Fatalf("\nshould have error, but got: %+v", err)
	}

	if resp.StatusCode != 500 {
		t.Fatalf("\nexpect status code of 500, instead of %d", resp.StatusCode)
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
	ctx context.Context, response interface{}, resp events.APIGatewayProxyResponse,
) (events.APIGatewayProxyResponse, error) {
	mode, ok := ctx.Value(KeyEncMode).(string)
	fmt.Println(mode)
	if ok && mode == "fail_encode" {
		return resp, fmt.Errorf("fail encoding")
	}

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

func makeTest01FailEndpoint(_ serviceTest01) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		return nil, fmt.Errorf("test error endpoint")
	}
}

type serviceTest01 struct{}

func (ts *serviceTest01) hello(name string) string {
	return fmt.Sprintf("hello %s", name)
}
