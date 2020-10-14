package awssqs

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport"
)

// Consumer wraps an endpoint and provides a handler for SQS messages.
type Consumer struct {
	sqsClient    sqsiface.SQSAPI
	e            endpoint.Endpoint
	dec          DecodeRequestFunc
	enc          EncodeResponseFunc
	wantRep      WantReplyFunc
	queueURL     string
	before       []ConsumerRequestFunc
	after        []ConsumerResponseFunc
	errorEncoder ErrorEncoder
	finalizer    []ConsumerFinalizerFunc
	errorHandler transport.ErrorHandler
}

// NewConsumer constructs a new Consumer, which provides a Consume method
// and message handlers that wrap the provided endpoint.
func NewConsumer(
	sqsClient sqsiface.SQSAPI,
	e endpoint.Endpoint,
	dec DecodeRequestFunc,
	enc EncodeResponseFunc,
	queueURL string,
	options ...ConsumerOption,
) *Consumer {
	s := &Consumer{
		sqsClient:    sqsClient,
		e:            e,
		dec:          dec,
		enc:          enc,
		wantRep:      DoNotRespond,
		queueURL:     queueURL,
		errorEncoder: DefaultErrorEncoder,
		errorHandler: transport.NewLogErrorHandler(log.NewNopLogger()),
	}
	for _, option := range options {
		option(s)
	}
	return s
}

// ConsumerOption sets an optional parameter for consumers.
type ConsumerOption func(*Consumer)

// ConsumerBefore functions are executed on the producer request object before the
// request is decoded.
func ConsumerBefore(before ...ConsumerRequestFunc) ConsumerOption {
	return func(c *Consumer) { c.before = append(c.before, before...) }
}

// ConsumerAfter functions are executed on the consumer reply after the
// endpoint is invoked, but before anything is published to the reply.
func ConsumerAfter(after ...ConsumerResponseFunc) ConsumerOption {
	return func(c *Consumer) { c.after = append(c.after, after...) }
}

// ConsumerErrorEncoder is used to encode errors to the consumer reply
// whenever they're encountered in the processing of a request. Clients can
// use this to provide custom error formatting. By default,
// errors will be published with the DefaultErrorEncoder.
func ConsumerErrorEncoder(ee ErrorEncoder) ConsumerOption {
	return func(c *Consumer) { c.errorEncoder = ee }
}

// ConsumerWantReplyFunc overrides the default value for the consumer's
// wantRep field.
func ConsumerWantReplyFunc(replyFunc WantReplyFunc) ConsumerOption {
	return func(c *Consumer) { c.wantRep = replyFunc }
}

// ConsumerErrorHandler is used to handle non-terminal errors. By default, non-terminal errors
// are ignored. This is intended as a diagnostic measure. Finer-grained control
// of error handling, including logging in more detail, should be performed in a
// custom ConsumerErrorEncoder which has access to the context.
func ConsumerErrorHandler(errorHandler transport.ErrorHandler) ConsumerOption {
	return func(c *Consumer) { c.errorHandler = errorHandler }
}

// ConsumerFinalizer is executed once all the received SQS messages are done being processed.
// By default, no finalizer is registered.
func ConsumerFinalizer(f ...ConsumerFinalizerFunc) ConsumerOption {
	return func(c *Consumer) { c.finalizer = f }
}

// ConsumerDeleteMessageBefore returns a ConsumerOption that appends a function
// that delete the message from queue to the list of consumer's before functions.
func ConsumerDeleteMessageBefore() ConsumerOption {
	return func(c *Consumer) {
		deleteBefore := func(ctx context.Context, cancel context.CancelFunc, msg *sqs.Message) context.Context {
			if err := deleteMessage(ctx, c.sqsClient, c.queueURL, msg); err != nil {
				c.errorHandler.Handle(ctx, err)
				c.errorEncoder(ctx, err, msg, c.sqsClient)
				cancel()
			}
			return ctx
		}
		c.before = append(c.before, deleteBefore)
	}
}

// ConsumerDeleteMessageAfter returns a ConsumerOption that appends a function
// that delete a message from queue to the list of consumer's after functions.
func ConsumerDeleteMessageAfter() ConsumerOption {
	return func(c *Consumer) {
		deleteAfter := func(ctx context.Context, cancel context.CancelFunc, msg *sqs.Message, _ *sqs.SendMessageInput) context.Context {
			if err := deleteMessage(ctx, c.sqsClient, c.queueURL, msg); err != nil {
				c.errorHandler.Handle(ctx, err)
				c.errorEncoder(ctx, err, msg, c.sqsClient)
				cancel()
			}
			return ctx
		}
		c.after = append(c.after, deleteAfter)
	}
}

// ServeMessage serves an SQS message.
func (c Consumer) ServeMessage(ctx context.Context, msg *sqs.Message) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if len(c.finalizer) > 0 {
		defer func() {
			for _, f := range c.finalizer {
				f(ctx, msg)
			}
		}()
	}

	for _, f := range c.before {
		ctx = f(ctx, cancel, msg)
	}

	req, err := c.dec(ctx, msg)
	if err != nil {
		c.errorHandler.Handle(ctx, err)
		c.errorEncoder(ctx, err, msg, c.sqsClient)
		return err
	}

	response, err := c.e(ctx, req)
	if err != nil {
		c.errorHandler.Handle(ctx, err)
		c.errorEncoder(ctx, err, msg, c.sqsClient)
		return err
	}

	responseMsg := sqs.SendMessageInput{}
	for _, f := range c.after {
		ctx = f(ctx, cancel, msg, &responseMsg)
	}

	if !c.wantRep(ctx, msg) {
		return nil
	}

	if err := c.enc(ctx, &responseMsg, response); err != nil {
		c.errorHandler.Handle(ctx, err)
		c.errorEncoder(ctx, err, msg, c.sqsClient)
		return err
	}

	if _, err := c.sqsClient.SendMessageWithContext(ctx, &responseMsg); err != nil {
		c.errorHandler.Handle(ctx, err)
		c.errorEncoder(ctx, err, msg, c.sqsClient)
		return err
	}
	return nil
}

// ErrorEncoder is responsible for encoding an error to the consumer's reply.
// Users are encouraged to use custom ErrorEncoders to encode errors to
// their replies, and will likely want to pass and check for their own error
// types.
type ErrorEncoder func(ctx context.Context, err error, req *sqs.Message, sqsClient sqsiface.SQSAPI)

// ConsumerFinalizerFunc can be used to perform work at the end of a request
// from a producer, after the response has been written to the producer. The
// principal intended use is for request logging.
// Can also be used to delete messages once fully proccessed.
type ConsumerFinalizerFunc func(ctx context.Context, msg *sqs.Message)

// WantReplyFunc encapsulates logic to check whether message awaits response or not
// for example check for a given message attribute value.
type WantReplyFunc func(context.Context, *sqs.Message) bool

// DefaultErrorEncoder simply ignores the message. It does not reply.
func DefaultErrorEncoder(context.Context, error, *sqs.Message, sqsiface.SQSAPI) {
}

func deleteMessage(ctx context.Context, sqsClient sqsiface.SQSAPI, queueURL string, msg *sqs.Message) error {
	_, err := sqsClient.DeleteMessageWithContext(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      &queueURL,
		ReceiptHandle: msg.ReceiptHandle,
	})
	return err
}

// DoNotRespond is a WantReplyFunc and is the default value for consumer's wantRep field.
// It indicates that the message do not expect a response.
func DoNotRespond(context.Context, *sqs.Message) bool {
	return false
}

// EncodeJSONResponse marshals response as json and loads it into an sqs.SendMessageInput MessageBody.
func EncodeJSONResponse(_ context.Context, input *sqs.SendMessageInput, response interface{}) error {
	payload, err := json.Marshal(response)
	if err != nil {
		return err
	}
	input.MessageBody = aws.String(string(payload))
	return nil
}
