package awssqs

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/go-kit/kit/endpoint"
)

type contextKey int

const (
	// ContextKeyResponseQueueURL is the context key that allows fetching
	// the response queue URL from context
	ContextKeyResponseQueueURL contextKey = iota
)

// Producer wraps an SQS client and queue, and provides a method that
// implements endpoint.Endpoint.
type Producer struct {
	sqsClient sqsiface.SQSAPI
	queueURL  string
	enc       EncodeRequestFunc
	dec       DecodeResponseFunc
	before    []ProducerRequestFunc
	after     []ProducerResponseFunc
	timeout   time.Duration
}

// NewProducer constructs a usable Producer for a single remote method.
func NewProducer(
	sqsClient sqsiface.SQSAPI,
	queueURL string,
	enc EncodeRequestFunc,
	dec DecodeResponseFunc,
	options ...ProducerOption,
) *Producer {
	p := &Producer{
		sqsClient: sqsClient,
		queueURL:  queueURL,
		enc:       enc,
		dec:       dec,
		timeout:   20 * time.Second,
	}
	for _, option := range options {
		option(p)
	}
	return p
}

// ProducerOption sets an optional parameter for clients.
type ProducerOption func(*Producer)

// ProducerBefore sets the RequestFuncs that are applied to the outgoing SQS
// request before it's invoked.
func ProducerBefore(before ...ProducerRequestFunc) ProducerOption {
	return func(p *Producer) { p.before = append(p.before, before...) }
}

// ProducerAfter sets the ClientResponseFuncs applied to the incoming SQS
// request prior to it being decoded. This is useful for obtaining the response
// and adding any information onto the context prior to decoding.
func ProducerAfter(after ...ProducerResponseFunc) ProducerOption {
	return func(p *Producer) { p.after = append(p.after, after...) }
}

// ProducerTimeout sets the available timeout for an SQS request.
func ProducerTimeout(timeout time.Duration) ProducerOption {
	return func(p *Producer) { p.timeout = timeout }
}

// SetProducerResponseQueueURL sets this as before or after function
func SetProducerResponseQueueURL(url string) ProducerRequestFunc {
	return func(ctx context.Context, _ *sqs.SendMessageInput) context.Context {
		return context.WithValue(ctx, ContextKeyResponseQueueURL, url)
	}
}

// Endpoint returns a usable endpoint that invokes the remote endpoint.
func (p Producer) Endpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		ctx, cancel := context.WithTimeout(ctx, p.timeout)
		defer cancel()
		msgInput := sqs.SendMessageInput{
			QueueUrl: &p.queueURL,
		}
		if err := p.enc(ctx, &msgInput, request); err != nil {
			return nil, err
		}

		for _, f := range p.before {
			ctx = f(ctx, &msgInput)
		}

		output, err := p.sqsClient.SendMessageWithContext(ctx, &msgInput)
		if err != nil {
			return nil, err
		}

		var responseMsg *sqs.Message
		for _, f := range p.after {
			ctx, responseMsg, err = f(ctx, p.sqsClient, output)
			if err != nil {
				return nil, err
			}
		}

		response, err := p.dec(ctx, responseMsg)
		if err != nil {
			return nil, err
		}

		return response, nil
	}
}

// EncodeJSONRequest is an EncodeRequestFunc that serializes the request as a
// JSON object and loads it as the MessageBody of the sqs.SendMessageInput.
// This can be enough for most JSON over SQS communications.
func EncodeJSONRequest(_ context.Context, msg *sqs.SendMessageInput, request interface{}) error {
	b, err := json.Marshal(request)
	if err != nil {
		return err
	}

	msg.MessageBody = aws.String(string(b))

	return nil
}

// NoResponseDecode is a DecodeResponseFunc that can be used when no response is needed.
// It returns nil value and nil error.
func NoResponseDecode(_ context.Context, _ *sqs.Message) (interface{}, error) {
	return nil, nil
}
