package amqp

import (
	"context"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/streadway/amqp"
)

// The golang AMQP implementation requires the []byte representation of
// correlation id strings to have a maximum length of 255 bytes.
const maxCorrelationIdLength = 255

// Publisher wraps an AMQP channel and queue, and provides a method that
// implements endpoint.Endpoint.
type Publisher struct {
	ch      Channel
	q       *amqp.Queue
	enc     EncodeRequestFunc
	dec     DecodeResponseFunc
	before  []RequestFunc
	after   []PublisherResponseFunc
	timeout time.Duration
}

// NewPublisher constructs a usable Publisher for a single remote method.
func NewPublisher(
	ch Channel,
	q *amqp.Queue,
	enc EncodeRequestFunc,
	dec DecodeResponseFunc,
	options ...PublisherOption,
) *Publisher {
	p := &Publisher{
		ch:      ch,
		q:       q,
		enc:     enc,
		dec:     dec,
		timeout: 10 * time.Second,
	}
	for _, option := range options {
		option(p)
	}
	return p
}

// PublisherOption sets an optional parameter for clients.
type PublisherOption func(*Publisher)

// PublisherBefore sets the RequestFuncs that are applied to the outgoing AMQP
// request before it's invoked.
func PublisherBefore(before ...RequestFunc) PublisherOption {
	return func(p *Publisher) { p.before = append(p.before, before...) }
}

// PublisherAfter sets the ClientResponseFuncs applied to the incoming AMQP
// request prior to it being decoded. This is useful for obtaining anything off
// of the response and adding onto the context prior to decoding.
func PublisherAfter(after ...PublisherResponseFunc) PublisherOption {
	return func(p *Publisher) { p.after = append(p.after, after...) }
}

// PublisherTimeout sets the available timeout for an AMQP request.
func PublisherTimeout(timeout time.Duration) PublisherOption {
	return func(p *Publisher) { p.timeout = timeout }
}

// Endpoint returns a usable endpoint that invokes the remote endpoint.
func (p Publisher) Endpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		ctx, cancel := context.WithTimeout(ctx, p.timeout)
		defer cancel()

		pub := amqp.Publishing{
			ReplyTo:       p.q.Name,
			CorrelationId: randomString(randInt(5, maxCorrelationIdLength)),
		}

		if err := p.enc(ctx, &pub, request); err != nil {
			return nil, err
		}

		for _, f := range p.before {
			ctx = f(ctx, &pub)
		}

		deliv, err := p.publishAndConsumeFirstMatchingResponse(ctx, &pub)
		if err != nil {
			return nil, err
		}

		for _, f := range p.after {
			ctx = f(ctx, deliv)
		}
		response, err := p.dec(ctx, deliv)
		if err != nil {
			return nil, err
		}

		return response, nil
	}
}

// publishAndConsumeFirstMatchingResponse publishes the specified Publishing
// and returns the first Delivery object with the matching correlationId.
// If the context times out while waiting for a reply, an error will be returned.
func (p Publisher) publishAndConsumeFirstMatchingResponse(
	ctx context.Context,
	pub *amqp.Publishing,
) (*amqp.Delivery, error) {
	err := p.ch.Publish(
		getPublishExchange(ctx),
		getPublishKey(ctx),
		false, //mandatory
		false, //immediate
		*pub,
	)
	if err != nil {
		return nil, err
	}
	autoAck := getConsumeAutoAck(ctx)

	msg, err := p.ch.Consume(
		p.q.Name,
		"", //consumer
		autoAck,
		false, //exclusive
		false, //noLocal
		false, //noWait
		getConsumeArgs(ctx),
	)
	if err != nil {
		return nil, err
	}

	for {
		select {
		case d := <-msg:
			if d.CorrelationId == pub.CorrelationId {
				if !autoAck {
					d.Ack(false) //multiple
				}
				return &d, nil
			}

		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

}
