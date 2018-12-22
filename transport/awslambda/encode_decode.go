package awslambda

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
)

// DecodeRequestFunc extracts a user-domain request object from an
// AWS API gateway event. One straightforward DecodeRequestFunc could be
// something that JSON decodes from the request body to the concrete request type.
type DecodeRequestFunc func(
	context.Context, events.APIGatewayProxyRequest,
) (request interface{}, err error)

// EncodeResponseFunc encodes the passed response object into
// API gateway proxy response format.
type EncodeResponseFunc func(
	ctx context.Context, response interface{}, resp events.APIGatewayProxyResponse,
) (events.APIGatewayProxyResponse, error)

// ErrorEncoder is responsible for encoding an error.
type ErrorEncoder func(
	ctx context.Context, err error, resp events.APIGatewayProxyResponse,
) (events.APIGatewayProxyResponse, error)
