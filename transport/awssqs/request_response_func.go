package awssqs

import (
	"context"

	"github.com/aws/aws-sdk-go/service/sqs"
)

// ConsumerRequestFunc may take information from a consumer request result and
// put it into a request context. In Consumers, RequestFuncs are executed prior
// to invoking the endpoint.
// use cases eg. in Consumer : extract message info to context, or sort messages
type ConsumerRequestFunc func(context.Context, *[]*sqs.Message) context.Context

// PublisherRequestFunc may take information from a publisher request and put it into a
// request context, or add some informations to SendMessageInput. In Publishers,
// RequestFuncs are executed prior to publishing the msg but after encoding.
// use cases eg. in Publisher : add message attributes to SendMessageInput
type PublisherRequestFunc func(context.Context, *sqs.SendMessageInput) context.Context

// ConsumerResponseFunc may take information from a request context and use it to
// manipulate a Publisher. ConsumerResponseFunc are only executed in
// consumers, after invoking the endpoint but prior to publishing a reply.
// eg. Pipe information from req message to response MessageInput or delete msg from queue
// Should also delete message from leftMsgs slice
type ConsumerResponseFunc func(context.Context, *sqs.Message, *sqs.SendMessageInput, *[]*sqs.Message) context.Context

// PublisherResponseFunc may take information from an sqs send message output and
// ask for response. SQS is not req-reply out-of-the-box. Response needs to be fetched.
// PublisherResponseFunc are only executed in publishers, after a request has been made,
// but prior to its resp being decoded. So this is the perfect place to fetch actual response.
type PublisherResponseFunc func(context.Context, *sqs.SendMessageOutput) (context.Context, *sqs.Message, error)

// WantReplyFunc encapsulates logic to check whether message awaits response or not
// eg. Check for a given attribute value
type WantReplyFunc func(context.Context, *sqs.Message) bool

// VisibilityTimeoutFunc encapsulates logic to extend messages visibility timeout
type VisibilityTimeoutFunc func(context.Context, SQSClient, *string, int64, *[]*sqs.Message) error
