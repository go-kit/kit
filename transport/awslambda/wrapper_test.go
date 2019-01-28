package awslambda

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestInvokeWithWrapperHappyPath(t *testing.T) {
	svc := serviceTest01{}

	helloHandler := NewHandler(
		makeTest01HelloEndpoint(svc),
		DecodeRequestWrapper(func(
			_ context.Context,
			apigwReq events.APIGatewayProxyRequest,
		) (helloRequest, error) {
			request := helloRequest{}
			err := json.Unmarshal([]byte(apigwReq.Body), &request)
			return request, err
		}),
		EncodeResponseWrapper(func(
			_ context.Context,
			response helloResponse,
		) (apigwResp events.APIGatewayProxyResponse, err error) {
			respByte, err := json.Marshal(response)
			if err != nil {
				return apigwResp, err
			}

			apigwResp.Body = string(respByte)
			apigwResp.StatusCode = 200
			return apigwResp, err
		}),
	)

	ctx := context.Background()
	req, _ := json.Marshal(events.APIGatewayProxyRequest{
		Body: `{"name":"john doe"}`,
	})
	resp, err := helloHandler.Invoke(ctx, req)

	if err != nil {
		t.Fatalf("should have no error, but got: %+v", err)
	}

	apigwResp := events.APIGatewayProxyResponse{}
	err = json.Unmarshal(resp, &apigwResp)
	if err != nil {
		t.Fatalf("Should have no error, but got: %+v", err)
	}

	response := helloResponse{}
	err = json.Unmarshal([]byte(apigwResp.Body), &response)
	if err != nil {
		t.Fatalf("Should have no error, but got: %+v", err)
	}

	expectedGreeting := "hello john doe"
	if response.Greeting != expectedGreeting {
		t.Fatalf(
			"Expect: %s, Actual: %s", expectedGreeting, response.Greeting)
	}
}

func TestInvokeWithWrapperErrorEncoder(t *testing.T) {
	svc := serviceTest01{}

	helloHandler := NewHandler(
		makeTest01HelloEndpoint(svc),
		DecodeRequestWrapper(func(
			_ context.Context,
			apigwReq events.APIGatewayProxyRequest,
		) (helloRequest, error) {
			request := helloRequest{}
			err := json.Unmarshal([]byte(apigwReq.Body), &request)
			return request, err
		}),
		EncodeResponseWrapper(func(
			_ context.Context,
			response helloResponse,
		) (apigwResp events.APIGatewayProxyResponse, err error) {
			respByte, err := json.Marshal(response)
			if err != nil {
				return apigwResp, err
			}

			apigwResp.Body = string(respByte)
			apigwResp.StatusCode = 200
			return apigwResp, err
		}),
		HandlerErrorEncoder(ErrorEncoderWrapper(func(
			_ context.Context,
			err error,
		) (apigwResp events.APIGatewayProxyResponse, returnErr error) {
			apigwResp.Body = `{"error":"yes"}`
			apigwResp.StatusCode = 500
			return apigwResp, nil
		})),
	)

	ctx := context.Background()
	req, _ := json.Marshal(events.APIGatewayProxyRequest{
		Body: "<xml>",
	})
	resp, err := helloHandler.Invoke(ctx, req)

	if err != nil {
		t.Fatalf("Should have no error, but got: %+v", err)
	}

	apigwResp := events.APIGatewayProxyResponse{}
	json.Unmarshal(resp, &apigwResp)
	if apigwResp.StatusCode != 500 {
		t.Fatalf("Expect status code of 500, instead of %d", apigwResp.StatusCode)
	}
}

func TestInvalidDecodeRequestWrapper(t *testing.T) {
	svc := serviceTest01{}
	validEncodeResponse := EncodeResponseWrapper(func(
		_ context.Context,
		response helloResponse,
	) (apigwResp events.APIGatewayProxyResponse, err error) {
		respByte, err := json.Marshal(response)
		if err != nil {
			return apigwResp, err
		}

		apigwResp.Body = string(respByte)
		apigwResp.StatusCode = 200
		return apigwResp, err
	})

	testCases := []struct {
		decoder        interface{}
		expectedErrMsg string
	}{
		{
			decoder:        nil,
			expectedErrMsg: "decoder is nil",
		},
		{
			decoder:        "hello",
			expectedErrMsg: "decoder kind string is not func",
		},
		{
			decoder:        func() {},
			expectedErrMsg: "decoder must take two arguments, but it takes 0",
		},
		{
			decoder:        func(s string, b string) {},
			expectedErrMsg: "decoder takes two arguments, but the first is not Context. got string",
		},
		{
			decoder: func(
				ctx context.Context,
				req events.APIGatewayProxyRequest,
			) {
			},
			expectedErrMsg: "decoder must return two values",
		},
		{
			decoder: func(
				ctx context.Context, req events.APIGatewayProxyRequest,
			) (helloRequest, string) {
				request := helloRequest{}
				return request, "yes"
			},
			expectedErrMsg: "decoder returns two values, but the second does not implement error",
		},
	}

	for _, tc := range testCases {
		helloHandler := NewHandler(
			makeTest01HelloEndpoint(svc),
			DecodeRequestWrapper(tc.decoder),
			validEncodeResponse,
		)

		ctx := context.Background()
		req, _ := json.Marshal(events.APIGatewayProxyRequest{
			Body: `{"name":"john doe"}`,
		})

		_, err := helloHandler.Invoke(ctx, req)

		if err == nil {
			t.Errorf("Should have error")
		}
		if err.Error() != tc.expectedErrMsg {
			t.Fatalf(
				"Expected: %+v Actual: %+v",
				tc.expectedErrMsg, err.Error())
		}
	}
}

func TestInvalidEncodeResponseWrapper(t *testing.T) {
	svc := serviceTest01{}
	validDecoder := DecodeRequestWrapper(func(
		_ context.Context,
		apigwReq events.APIGatewayProxyRequest,
	) (helloRequest, error) {
		request := helloRequest{}
		err := json.Unmarshal([]byte(apigwReq.Body), &request)
		return request, err
	})

	testCases := []struct {
		encoder        interface{}
		expectedErrMsg string
	}{
		{
			encoder:        nil,
			expectedErrMsg: "encoder is nil",
		},
		{
			encoder:        "hello",
			expectedErrMsg: "encoder kind string is not func",
		},
		{
			encoder:        func() {},
			expectedErrMsg: "encoder must take two arguments, but it takes 0",
		},
		{
			encoder:        func(s string, b string) {},
			expectedErrMsg: "encoder takes two arguments, but the first is not Context. got string",
		},
		{
			encoder: func(
				ctx context.Context, response helloResponse,
			) {
			},
			expectedErrMsg: "encoder must return two values",
		},
		{
			encoder: func(
				ctx context.Context,
				response helloResponse,
			) (apigwResp events.APIGatewayProxyResponse, s string) {
				respByte, err := json.Marshal(response)
				if err != nil {
					return apigwResp, "err"
				}

				apigwResp.Body = string(respByte)
				apigwResp.StatusCode = 200
				return apigwResp, "err"
			},
			expectedErrMsg: "encoder returns two values, but the second does not implement error",
		},
	}

	for _, tc := range testCases {
		helloHandler := NewHandler(
			makeTest01HelloEndpoint(svc),
			validDecoder,
			EncodeResponseWrapper(tc.encoder),
		)

		ctx := context.Background()
		req, _ := json.Marshal(events.APIGatewayProxyRequest{
			Body: `{"name":"john doe"}`,
		})

		_, err := helloHandler.Invoke(ctx, req)

		if err == nil {
			t.Errorf("Should have error")
		}
		if err.Error() != tc.expectedErrMsg {
			t.Fatalf(
				"Expected: %+v Actual: %+v",
				tc.expectedErrMsg, err.Error())
		}
	}
}

func TestInvalidErrorEncoderWrapper(t *testing.T) {
	svc := serviceTest01{}
	validDecoder := DecodeRequestWrapper(func(
		_ context.Context,
		apigwReq events.APIGatewayProxyRequest,
	) (helloRequest, error) {
		request := helloRequest{}
		err := json.Unmarshal([]byte(apigwReq.Body), &request)
		return request, err
	})
	validEncoder := EncodeResponseWrapper(func(
		_ context.Context,
		response helloResponse,
	) (apigwResp events.APIGatewayProxyResponse, err error) {
		respByte, err := json.Marshal(response)
		if err != nil {
			return apigwResp, err
		}

		apigwResp.Body = string(respByte)
		apigwResp.StatusCode = 200
		return apigwResp, err
	})

	testCases := []struct {
		errorEncoder   interface{}
		expectedErrMsg string
	}{
		{
			errorEncoder:   nil,
			expectedErrMsg: "errorEncoder is nil",
		},
		{
			errorEncoder:   "hello",
			expectedErrMsg: "errorEncoder kind string is not func",
		},
		{
			errorEncoder:   func() {},
			expectedErrMsg: "errorEncoder must take two arguments, but it takes 0",
		},
		{
			errorEncoder:   func(s string, b string) {},
			expectedErrMsg: "errorEncoder takes two arguments, but the first is not Context. got string",
		},
		{
			errorEncoder: func(
				ctx context.Context,
				b string,
			) {
			},
			expectedErrMsg: "errorEncoder takes two arguments, but the second is not error. got string",
		},
		{
			errorEncoder: func(
				ctx context.Context,
				err error,
			) {
			},
			expectedErrMsg: "errorEncoder must return two values",
		},
		{
			errorEncoder: func(
				ctx context.Context,
				err error,
			) (apigwResp events.APIGatewayProxyResponse, s string) {
				apigwResp.Body = `{"error":"yes"}`
				apigwResp.StatusCode = 500
				return apigwResp, "nil"
			},
			expectedErrMsg: "errorEncoder returns two values, but the second does not implement error",
		},
	}

	for _, tc := range testCases {
		helloHandler := NewHandler(
			makeTest01HelloEndpoint(svc),
			validDecoder,
			validEncoder,
			HandlerErrorEncoder(ErrorEncoderWrapper(tc.errorEncoder)),
		)

		ctx := context.Background()
		req, _ := json.Marshal(events.APIGatewayProxyRequest{
			Body: `<err-format>`,
		})

		_, err := helloHandler.Invoke(ctx, req)

		if err == nil {
			t.Errorf("Should have error")
		}
		if err.Error() != tc.expectedErrMsg {
			t.Fatalf(
				"Expected: %+v Actual: %+v",
				tc.expectedErrMsg, err.Error())
		}
	}
}

func TestWrapperInvalidPayloadFormat(t *testing.T) {
	svc := serviceTest01{}

	helloHandler := NewHandler(
		makeTest01HelloEndpoint(svc),
		DecodeRequestWrapper(func(
			_ context.Context,
			apigwReq events.APIGatewayProxyRequest,
		) (helloRequest, error) {
			request := helloRequest{}
			err := json.Unmarshal([]byte(apigwReq.Body), &request)
			return request, err
		}),
		EncodeResponseWrapper(func(
			_ context.Context,
			response helloResponse,
		) (apigwResp events.APIGatewayProxyResponse, err error) {
			respByte, err := json.Marshal(response)
			if err != nil {
				return apigwResp, err
			}

			apigwResp.Body = string(respByte)
			apigwResp.StatusCode = 200
			return apigwResp, err
		}),
	)

	ctx := context.Background()
	req := []byte("<invalid-format />")
	_, err := helloHandler.Invoke(ctx, req)

	if err == nil {
		t.Fatalf("Should have error")
	}
}

func TestWrapperErrorInEncodeResponse(t *testing.T) {
	svc := serviceTest01{}

	helloHandler := NewHandler(
		makeTest01HelloEndpoint(svc),
		DecodeRequestWrapper(func(
			_ context.Context,
			apigwReq events.APIGatewayProxyRequest,
		) (helloRequest, error) {
			request := helloRequest{}
			err := json.Unmarshal([]byte(apigwReq.Body), &request)
			return request, err
		}),
		EncodeResponseWrapper(func(
			_ context.Context,
			response helloResponse,
		) (apigwResp events.APIGatewayProxyResponse, err error) {
			return apigwResp, fmt.Errorf("error")
		}),
	)

	ctx := context.Background()
	req, _ := json.Marshal(events.APIGatewayProxyRequest{
		Body: `{"name":"john doe"}`,
	})
	_, err := helloHandler.Invoke(ctx, req)

	if err == nil {
		t.Fatalf("Should have error")
	}
}

func TestInvokeWithWrapperErrorEncoderReturnsError(t *testing.T) {
	svc := serviceTest01{}

	helloHandler := NewHandler(
		makeTest01HelloEndpoint(svc),
		DecodeRequestWrapper(func(
			_ context.Context,
			apigwReq events.APIGatewayProxyRequest,
		) (helloRequest, error) {
			request := helloRequest{}
			err := json.Unmarshal([]byte(apigwReq.Body), &request)
			return request, err
		}),
		EncodeResponseWrapper(func(
			_ context.Context,
			response helloResponse,
		) (apigwResp events.APIGatewayProxyResponse, err error) {
			respByte, err := json.Marshal(response)
			if err != nil {
				return apigwResp, err
			}

			apigwResp.Body = string(respByte)
			apigwResp.StatusCode = 200
			return apigwResp, err
		}),
		HandlerErrorEncoder(ErrorEncoderWrapper(func(
			_ context.Context,
			err error,
		) (apigwResp events.APIGatewayProxyResponse, returnErr error) {
			return apigwResp, fmt.Errorf("error")
		})),
	)

	ctx := context.Background()
	req, _ := json.Marshal(events.APIGatewayProxyRequest{
		Body: "<xml>",
	})
	_, err := helloHandler.Invoke(ctx, req)

	if err == nil {
		t.Fatalf("Should have error")
	}
}
