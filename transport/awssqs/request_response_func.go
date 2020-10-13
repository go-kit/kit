package awssqs

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
)

// ConsumerRequestFunc may take information from a consumer request result and
// put it into a request context. In Consumers, RequestFuncs are executed prior
// to invoking the endpoint.
// use cases eg. in Consumer : extract message into context, or filter received messages.
type ConsumerRequestFunc func(context.Context, *[]*sqs.Message) context.Context

// ProducerRequestFunc may take information from a producer request and put it into a
// request context, or add some informations to SendMessageInput. In Producers,
// RequestFuncs are executed prior to publishing the message but after encoding.
// use cases eg. in Producer : enforce some message attributes to SendMessageInput.
type ProducerRequestFunc func(ctx context.Context, input *sqs.SendMessageInput) context.Context

// ConsumerResponseFunc may take information from a request context and use it to
// manipulate a Producer. ConsumerResponseFunc are only executed in
// consumers, after invoking the endpoint but prior to publishing a reply.
// use cases eg. : Pipe information from request message to response MessageInput,
// delete msg from queue or update leftMsgs slice.
type ConsumerResponseFunc func(context.Context, *sqs.Message, *sqs.SendMessageInput, *[]*sqs.Message, *sync.Mutex) context.Context

// ProducerResponseFunc may take information from an sqs.SendMessageOutput and
// fetch response using the Client. SQS is not req-reply out-of-the-box. Responses need to be fetched.
// ProducerResponseFunc are only executed in producers, after a request has been made,
// but prior to its response being decoded. So this is the perfect place to fetch actual response.
type ProducerResponseFunc func(context.Context, sqsiface.SQSAPI, *sqs.SendMessageOutput) (context.Context, *sqs.Message, error)
