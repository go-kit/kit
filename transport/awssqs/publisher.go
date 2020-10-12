package awssqs

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/go-kit/kit/endpoint"
)

// Publisher wraps an sqs client and queue, and provides a method that
// implements endpoint.Endpoint.
type Publisher struct {
	sqsClient        Client
	queueURL         string
	responseQueueURL string
	enc              EncodeRequestFunc
	dec              DecodeResponseFunc
	before           []PublisherRequestFunc
	after            []PublisherResponseFunc
	timeout          time.Duration
}

// NewPublisher constructs a usable Publisher for a single remote method.
func NewPublisher(
	sqsClient Client,
	queueURL string,
	responseQueueURL string,
	enc EncodeRequestFunc,
	dec DecodeResponseFunc,
	options ...PublisherOption,
) *Publisher {
	p := &Publisher{
		sqsClient:        sqsClient,
		queueURL:         queueURL,
		responseQueueURL: responseQueueURL,
		enc:              enc,
		dec:              dec,
		timeout:          20 * time.Second,
	}
	for _, option := range options {
		option(p)
	}
	return p
}

// PublisherOption sets an optional parameter for clients.
type PublisherOption func(*Publisher)

// PublisherBefore sets the RequestFuncs that are applied to the outgoing sqs
// request before it's invoked.
func PublisherBefore(before ...PublisherRequestFunc) PublisherOption {
	return func(p *Publisher) { p.before = append(p.before, before...) }
}

// PublisherAfter sets the ClientResponseFuncs applied to the incoming sqs
// request prior to it being decoded. This is useful for obtaining anything off
// of the response and adding onto the context prior to decoding.
func PublisherAfter(after ...PublisherResponseFunc) PublisherOption {
	return func(p *Publisher) { p.after = append(p.after, after...) }
}

// PublisherTimeout sets the available timeout for an sqs request.
func PublisherTimeout(timeout time.Duration) PublisherOption {
	return func(p *Publisher) { p.timeout = timeout }
}

// Endpoint returns a usable endpoint that invokes the remote endpoint.
func (p Publisher) Endpoint() endpoint.Endpoint {
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
			ctx = f(ctx, &msgInput, p.responseQueueURL)
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
