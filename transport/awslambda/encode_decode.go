package awslambda

import (
	"context"
)

// DecodeRequestFunc extracts a user-domain request object from an
// AWS Lambda payload.
type DecodeRequestFunc[Request any] func(context.Context, []byte) (Request, error)

// EncodeResponseFunc encodes the passed response object into []byte,
// ready to be sent as AWS Lambda response.
type EncodeResponseFunc[Response any] func(context.Context, Response) ([]byte, error)

// ErrorEncoder is responsible for encoding an error.
type ErrorEncoder func(ctx context.Context, err error) ([]byte, error)
