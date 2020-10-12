package awssqs

import (
	"context"

	"github.com/aws/aws-sdk-go/service/sqs"
)

// DecodeRequestFunc extracts a user-domain request object from
// an sqs message object. It is designed to be used in Consumers.
type DecodeRequestFunc func(context.Context, *sqs.Message) (request interface{}, err error)

// EncodeRequestFunc encodes the passed payload object into
// an sqs message object. It is designed to be used in Producers.
type EncodeRequestFunc func(context.Context, *sqs.SendMessageInput, interface{}) error

// EncodeResponseFunc encodes the passed response object to
// an sqs message object. It is designed to be used in Consumers.
type EncodeResponseFunc func(context.Context, *sqs.SendMessageInput, interface{}) error

// DecodeResponseFunc extracts a user-domain response object from
// an sqs message object. It is designed to be used in Producers.
type DecodeResponseFunc func(context.Context, *sqs.Message) (response interface{}, err error)
