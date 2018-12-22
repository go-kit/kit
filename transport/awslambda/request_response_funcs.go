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

// ServerResponseFunc may take information from a request context
// and use it to manipulate APIGatewayProxyResponse.
// ServerResponseFunc are only executed after invoking the endpoint
// but prior to returning a response.
type ServerResponseFunc func(context.Context, events.APIGatewayProxyResponse) context.Context

// ServerFinalizerFunc is executed at the end of every
// APIGatewayProxy request.
type ServerFinalizerFunc func(context.Context, events.APIGatewayProxyResponse, error)
