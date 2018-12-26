package awslambda

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
)

// DecodeRequestWrapper wraps the given decoderSymbol function, and generates
// the proper DecodeRequestFunc, based on the decoderSymbol function signature.
// The decoderSymbol function signature has to receive 2 args, which is the
// context.Context and event request to decode.
// It has also to return 2 values, the user-domain object and error.
func DecodeRequestWrapper(decoderSymbol interface{}) DecodeRequestFunc {
	if decoderSymbol == nil {
		return errorDecoderRequest(fmt.Errorf("decoder is nil"))
	}

	decoder := reflect.ValueOf(decoderSymbol)
	decoderType := reflect.TypeOf(decoderSymbol)
	if decoderType.Kind() != reflect.Func {
		return errorDecoderRequest(fmt.Errorf(
			"decoder kind %s is not %s", decoder.Kind(), reflect.Func))
	}

	if err := decoderValidateArguments(decoderType); err != nil {
		return errorDecoderRequest(err)
	}

	if err := decoderValidateReturns(decoderType); err != nil {
		return errorDecoderRequest(err)
	}

	return func(ctx context.Context, payload []byte) (interface{}, error) {
		// construct arguments
		var args []reflect.Value
		args = append(args, reflect.ValueOf(ctx))

		eventType := decoderType.In(decoderType.NumIn() - 1)
		event := reflect.New(eventType)

		if err := json.Unmarshal(payload, event.Interface()); err != nil {
			return nil, err
		}

		args = append(args, event.Elem())

		response := decoder.Call(args)

		// convert return values into (interface{}, error)
		var err error
		if len(response) > 0 {
			if errVal, ok := response[len(response)-1].Interface().(error); ok {
				err = errVal
			}
		}
		var val interface{}
		if len(response) > 1 {
			val = response[0].Interface()
		}

		return val, err
	}
}

func errorDecoderRequest(err error) DecodeRequestFunc {
	return func(context.Context, []byte) (interface{}, error) {
		return nil, err
	}
}

func decoderValidateArguments(decoderType reflect.Type) error {
	if decoderType.NumIn() != 2 {
		return fmt.Errorf(
			"decoder must take two arguments, but it takes %d",
			decoderType.NumIn())
	}

	contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
	argumentType := decoderType.In(0)
	if !argumentType.Implements(contextType) {
		return fmt.Errorf("decoder takes two arguments, but the first is not Context. got %s", argumentType.Kind())
	}

	return nil
}

func decoderValidateReturns(decoderType reflect.Type) error {
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if decoderType.NumOut() != 2 {
		return fmt.Errorf("decoder must return two values")
	}

	if !decoderType.Out(1).Implements(errorType) {
		return fmt.Errorf("decoder returns two values, but the second does not implement error")
	}
	return nil
}

// EncodeResponseWrapper wraps an encoder into EncoderResponseFunc.
// The encoderSymbol function has to take in 2 arguments. The first one is
// a context.Context, the second argument is a user-domain response
// object.
// The encoderSymbol function has also to return 2 values. The first one is
// the intended response event, the second value is about error.
// An example for first return value is event.APIGatewayProxyResponse.
func EncodeResponseWrapper(encoderSymbol interface{}) EncodeResponseFunc {
	if encoderSymbol == nil {
		return errorEncodeResponse(fmt.Errorf("encoder is nil"))
	}

	encoder := reflect.ValueOf(encoderSymbol)
	encoderType := reflect.TypeOf(encoderSymbol)
	if encoderType.Kind() != reflect.Func {
		return errorEncodeResponse(fmt.Errorf(
			"encoder kind %s is not %s", encoderType.Kind(), reflect.Func))
	}

	if err := encoderValidateArguments(encoderType); err != nil {
		return errorEncodeResponse(err)
	}

	if err := encoderValidateReturns(encoderType); err != nil {
		return errorEncodeResponse(err)
	}

	return func(ctx context.Context, response interface{}) ([]byte, error) {
		// construct arguments
		var args []reflect.Value
		args = append(args, reflect.ValueOf(ctx))
		args = append(args, reflect.ValueOf(response))

		rawResponse := encoder.Call(args)

		// convert return values into (interface{}, error)
		var err error
		if len(rawResponse) > 0 {
			if errVal, ok := rawResponse[len(rawResponse)-1].Interface().(error); ok {
				err = errVal
			}
		}
		var val interface{}
		if len(rawResponse) > 1 {
			val = rawResponse[0].Interface()
		}

		// convert return values into ([]byte, error)
		if err != nil {
			return nil, err
		}

		responseByte, err := json.Marshal(val)
		return responseByte, err
	}
}

func errorEncodeResponse(err error) EncodeResponseFunc {
	return func(context.Context, interface{}) ([]byte, error) {
		return nil, err
	}
}

func encoderValidateArguments(encoderType reflect.Type) error {
	if encoderType.NumIn() != 2 {
		return fmt.Errorf(
			"encoder must take two arguments, but it takes %d",
			encoderType.NumIn())
	}

	contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
	argumentType := encoderType.In(0)
	if !argumentType.Implements(contextType) {
		return fmt.Errorf("encoder takes two arguments, but the first is not Context. got %s", argumentType.Kind())
	}

	return nil
}

func encoderValidateReturns(encoderType reflect.Type) error {
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if encoderType.NumOut() != 2 {
		return fmt.Errorf("encoder must return two values")
	}

	if !encoderType.Out(1).Implements(errorType) {
		return fmt.Errorf("encoder returns two values, but the second does not implement error")
	}
	return nil
}

// ErrorEncoderWrapper wraps a errorEncoder into a ErrorEncoder.
// The errorEncoder function has to accept 2 arguments. The first one
// is context.Context, and the second one is error.
// The errorEncoder function has to return 2 values. The first one is
// the intended event response, and the second one is error.
func ErrorEncoderWrapper(errorEncoderSymbol interface{}) ErrorEncoder {
	if errorEncoderSymbol == nil {
		return errorErrorEncoder(fmt.Errorf("errorEncoder is nil"))
	}

	errorEncoder := reflect.ValueOf(errorEncoderSymbol)
	errorEncoderType := reflect.TypeOf(errorEncoderSymbol)
	if errorEncoderType.Kind() != reflect.Func {
		return errorErrorEncoder(fmt.Errorf(
			"errorEncoder kind %s is not %s", errorEncoderType.Kind(), reflect.Func))
	}

	if err := errorEncoderValidateArguments(errorEncoderType); err != nil {
		return errorErrorEncoder(err)
	}

	if err := errorEncoderValidateReturns(errorEncoderType); err != nil {
		return errorErrorEncoder(err)
	}

	return func(ctx context.Context, err error) ([]byte, error) {
		// construct arguments
		var args []reflect.Value
		args = append(args, reflect.ValueOf(ctx))
		args = append(args, reflect.ValueOf(err))

		rawResponse := errorEncoder.Call(args)

		// convert return values into (interface{}, error)
		var returnErr error
		if len(rawResponse) > 0 {
			if errVal, ok := rawResponse[len(rawResponse)-1].Interface().(error); ok {
				returnErr = errVal
			}
		}
		var val interface{}
		if len(rawResponse) > 1 {
			val = rawResponse[0].Interface()
		}

		// convert return values into ([]byte, error)
		if returnErr != nil {
			return nil, returnErr
		}

		responseByte, returnErr := json.Marshal(val)
		return responseByte, returnErr
	}
}

func errorErrorEncoder(err error) ErrorEncoder {
	return func(ctx context.Context, inErr error) ([]byte, error) {
		return nil, err
	}
}

func errorEncoderValidateArguments(errorEncoderType reflect.Type) error {
	if errorEncoderType.NumIn() != 2 {
		return fmt.Errorf(
			"errorEncoder must take two arguments, but it takes %d",
			errorEncoderType.NumIn())
	}

	contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
	argumentType := errorEncoderType.In(0)
	if !argumentType.Implements(contextType) {
		return fmt.Errorf("errorEncoder takes two arguments, but the first is not Context. got %s", argumentType.Kind())
	}

	errorType := reflect.TypeOf((*error)(nil)).Elem()
	argumentType = errorEncoderType.In(1)
	if !argumentType.Implements(errorType) {
		return fmt.Errorf("errorEncoder takes two arguments, but the second is not error. got %s", argumentType.Kind())
	}

	return nil
}

func errorEncoderValidateReturns(errorEncoderType reflect.Type) error {
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if errorEncoderType.NumOut() != 2 {
		return fmt.Errorf("errorEncoder must return two values")
	}

	if !errorEncoderType.Out(1).Implements(errorType) {
		return fmt.Errorf("errorEncoder returns two values, but the second does not implement error")
	}
	return nil
}
