package awslambda

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
)

// ServerRequestFunc may take information from the received
// AWS APIGatewayProxy request and use it to place items
// in the request scoped context. ServerRequestFuncs are executed
// prior to invoking the endpoint.
type ServerRequestFunc func(context.Context, events.APIGatewayProxyRequest) context.Context
