package awssqs

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport"
)

// Consumer wraps an endpoint and provides and provides a handler for sqs msgs
type Consumer struct {
	sqsClient             SQSClient
	e                     endpoint.Endpoint
	dec                   DecodeRequestFunc
	enc                   EncodeResponseFunc
	wantRep               WantReplyFunc
	queueURL              *string
	dlQueueURL            *string
	visibilityTimeout     int64
	visibilityTimeoutFunc VisibilityTimeoutFunc
	before                []ConsumerRequestFunc
	after                 []ConsumerResponseFunc
	errorEncoder          ErrorEncoder
	finalizer             []ConsumerFinalizerFunc
	errorHandler          transport.ErrorHandler
}

// NewConsumer constructs a new Consumer, which provides a Consume method
// and message handlers that wrap the provided endpoint.
func NewConsumer(
	sqsClient SQSClient,
	e endpoint.Endpoint,
	dec DecodeRequestFunc,
	enc EncodeResponseFunc,
	wantRep WantReplyFunc,
	queueURL *string,
	dlQueueURL *string,
	visibilityTimeout int64,
	options ...ConsumerOption,
) *Consumer {
	s := &Consumer{
		sqsClient:             sqsClient,
		e:                     e,
		dec:                   dec,
		enc:                   enc,
		wantRep:               wantRep,
		queueURL:              queueURL,
		dlQueueURL:            dlQueueURL,
		visibilityTimeout:     visibilityTimeout,
		visibilityTimeoutFunc: DoNotExtendVisibilityTimeout,
		errorEncoder:          DefaultErrorEncoder,
		errorHandler:          transport.NewLogErrorHandler(log.NewNopLogger()),
	}
	for _, option := range options {
		option(s)
	}
	return s
}

// ConsumerOption sets an optional parameter for consumers.
type ConsumerOption func(*Consumer)

// ConsumerBefore functions are executed on the publisher request object before the
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

// ConsumerVisbilityTimeOutFunc is used to extend the visibility timeout
// for messages during when processing them. Clients can
// use this to provide custom visibility timeout extension. By default,
// visibility timeout are not extend.
func ConsumerVisbilityTimeOutFunc(vtFunc VisibilityTimeoutFunc) ConsumerOption {
	return func(c *Consumer) { c.visibilityTimeoutFunc = vtFunc }
}

// ConsumerErrorLogger is used to log non-terminal errors. By default, no errors
// are logged. This is intended as a diagnostic measure. Finer-grained control
// of error handling, including logging in more detail, should be performed in a
// custom ConsumerErrorEncoder which has access to the context.
// Deprecated: Use ConsumerErrorHandler instead.
func ConsumerErrorLogger(logger log.Logger) ConsumerOption {
	return func(c *Consumer) { c.errorHandler = transport.NewLogErrorHandler(logger) }
}

// ConsumerErrorHandler is used to handle non-terminal errors. By default, non-terminal errors
// are ignored. This is intended as a diagnostic measure. Finer-grained control
// of error handling, including logging in more detail, should be performed in a
// custom ConsumerErrorEncoder which has access to the context.
func ConsumerErrorHandler(errorHandler transport.ErrorHandler) ConsumerOption {
	return func(c *Consumer) { c.errorHandler = errorHandler }
}

// ConsumerFinalizer is executed at the end of every request from a publisher through SQS.
// By default, no finalizer is registered.
func ConsumerFinalizer(f ...ConsumerFinalizerFunc) ConsumerOption {
	return func(c *Consumer) { c.finalizer = f }
}

// Consume calls ReceiveMessageWithContext and handles messages
// having receiveMsgInput as param allows each user to have his own receive config
func (c Consumer) Consume(ctx context.Context, receiveMsgInput *sqs.ReceiveMessageInput) error {
	receiveMsgInput.QueueUrl = c.queueURL
	out, err := c.sqsClient.ReceiveMessageWithContext(ctx, receiveMsgInput)
	if err != nil {
		return err
	}
	return c.HandleMessages(ctx, out.Messages)
}

// HandleMessages handles the consumed messages
func (c Consumer) HandleMessages(ctx context.Context, msgs []*sqs.Message) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Copy msgs slice in leftMsgs
	leftMsgs := []*sqs.Message{}
	leftMsgs = append(leftMsgs, msgs...)

	// this func allows us to extend visibility timeout to give use
	// time to process the messages in leftMsgs
	go c.visibilityTimeoutFunc(ctx, c.sqsClient, c.queueURL, c.visibilityTimeout, &leftMsgs)

	if len(c.finalizer) > 0 {
		defer func() {
			for _, f := range c.finalizer {
				f(ctx, &msgs)
			}
		}()
	}

	for _, f := range c.before {
		ctx = f(ctx, &msgs)
	}

	for _, msg := range msgs {
		if err := c.HandleSingleMessage(ctx, msg, &leftMsgs); err != nil {
			return err
		}
	}
	return nil
}

// HandleSingleMessage handles a single sqs message
func (c Consumer) HandleSingleMessage(ctx context.Context, msg *sqs.Message, leftMsgs *[]*sqs.Message) error {
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
		ctx = f(ctx, msg, &responseMsg, leftMsgs)
	}

	if !c.wantRep(ctx, msg) {
		// Message does not expect answer
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

// ErrorEncoder is responsible for encoding an error to the consumer reply.
// Users are encouraged to use custom ErrorEncoders to encode errors to
// their replies, and will likely want to pass and check for their own error
// types.
type ErrorEncoder func(ctx context.Context, err error, req *sqs.Message, sqsClient SQSClient)

// ConsumerFinalizerFunc can be used to perform work at the end of a request
// from a publisher, after the response has been written to the publisher. The
// principal intended use is for request logging.
// Can also be used to delete messages once fully proccessed
type ConsumerFinalizerFunc func(ctx context.Context, msg *[]*sqs.Message)

// DefaultErrorEncoder simply ignores the message. It does not reply
// nor Ack/Nack the message.
func DefaultErrorEncoder(context.Context, error, *sqs.Message, SQSClient) {
}

// DoNotExtendVisibilityTimeout is the default value for visibilityTimeoutFunc
// It returns no error and does nothing
func DoNotExtendVisibilityTimeout(context.Context, SQSClient, *string, int64, *[]*sqs.Message) error {
	return nil
}

// EncodeJSONResponse marshals response as json and loads it into input MessageBody
func EncodeJSONResponse(_ context.Context, input *sqs.SendMessageInput, response interface{}) error {
	payload, err := json.Marshal(response)
	if err != nil {
		return err
	}
	input.MessageBody = aws.String(string(payload))
	return nil
}

// SQSClient is an interface to make testing possible.
// It is highly recommended to use *sqs.SQS as the interface implementation.
type SQSClient interface {
	SendMessageWithContext(ctx context.Context, input *sqs.SendMessageInput, opts ...request.Option) (*sqs.SendMessageOutput, error)
	ReceiveMessageWithContext(ctx context.Context, input *sqs.ReceiveMessageInput, opts ...request.Option) (*sqs.ReceiveMessageOutput, error)
	ChangeMessageVisibilityWithContext(ctx aws.Context, input *sqs.ChangeMessageVisibilityInput, opts ...request.Option) (*sqs.ChangeMessageVisibilityOutput, error)
}
