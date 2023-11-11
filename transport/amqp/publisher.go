package amqp

import (
	"context"
	"time"

	"github.com/openmesh/kit/endpoint"
	amqp "github.com/rabbitmq/amqp091-go"
)

// The golang AMQP implementation requires the []byte representation of
// correlation id strings to have a maximum length of 255 bytes.
const maxCorrelationIdLength = 255

// Publisher wraps an AMQP channel and queue, and provides a method that
// implements endpoint.Endpoint.
type Publisher[Request, Response any] struct {
	ch        Channel
	q         *amqp.Queue
	enc       EncodeRequestFunc[Request]
	dec       DecodeResponseFunc[Response]
	before    []RequestFunc
	after     []PublisherResponseFunc
	deliverer Deliverer[Request, Response]
	timeout   time.Duration
}

// NewPublisher constructs a usable Publisher for a single remote method.
func NewPublisher[Request, Response any](
	ch Channel,
	q *amqp.Queue,
	enc EncodeRequestFunc[Request],
	dec DecodeResponseFunc[Response],
	options ...PublisherOption[Request, Response],
) *Publisher[Request, Response] {
	p := &Publisher[Request, Response]{
		ch:        ch,
		q:         q,
		enc:       enc,
		dec:       dec,
		deliverer: DefaultDeliverer[Request, Response],
		timeout:   10 * time.Second,
	}
	for _, option := range options {
		option(p)
	}
	return p
}

// PublisherOption sets an optional parameter for clients.
type PublisherOption[Request, Response any] func(*Publisher[Request, Response])

// PublisherBefore sets the RequestFuncs that are applied to the outgoing AMQP
// request before it's invoked.
func PublisherBefore[Request, Response any](before ...RequestFunc) PublisherOption[Request, Response] {
	return func(p *Publisher[Request, Response]) { p.before = append(p.before, before...) }
}

// PublisherAfter sets the ClientResponseFuncs applied to the incoming AMQP
// request prior to it being decoded. This is useful for obtaining anything off
// of the response and adding onto the context prior to decoding.
func PublisherAfter[Request, Response any](after ...PublisherResponseFunc) PublisherOption[Request, Response] {
	return func(p *Publisher[Request, Response]) { p.after = append(p.after, after...) }
}

// PublisherDeliverer sets the deliverer function that the Publisher invokes.
func PublisherDeliverer[Request, Response any](deliverer Deliverer[Request, Response]) PublisherOption[Request, Response] {
	return func(p *Publisher[Request, Response]) { p.deliverer = deliverer }
}

// PublisherTimeout sets the available timeout for an AMQP request.
func PublisherTimeout[Request, Response any](timeout time.Duration) PublisherOption[Request, Response] {
	return func(p *Publisher[Request, Response]) { p.timeout = timeout }
}

// Endpoint returns a usable endpoint that invokes the remote endpoint.
func (p Publisher[Request, Response]) Endpoint() endpoint.Endpoint[Request, Response] {
	return func(ctx context.Context, request Request) (Response, error) {
		ctx, cancel := context.WithTimeout(ctx, p.timeout)
		defer cancel()

		pub := amqp.Publishing{
			ReplyTo:       p.q.Name,
			CorrelationId: randomString(randInt(5, maxCorrelationIdLength)),
		}

		if err := p.enc(ctx, &pub, request); err != nil {
			return *new(Response), err
		}

		for _, f := range p.before {
			// Affect only amqp.Publishing
			ctx = f(ctx, &pub, nil)
		}

		deliv, err := p.deliverer(ctx, p, &pub)
		if err != nil {
			return *new(Response), err
		}

		for _, f := range p.after {
			ctx = f(ctx, deliv)
		}
		response, err := p.dec(ctx, deliv)
		if err != nil {
			return *new(Response), err
		}

		return response, nil
	}
}

// Deliverer is invoked by the Publisher to publish the specified Publishing, and to
// retrieve the appropriate response Delivery object.
type Deliverer[Request, Response any] func(
	context.Context,
	Publisher[Request, Response],
	*amqp.Publishing,
) (*amqp.Delivery, error)

// DefaultDeliverer is a deliverer that publishes the specified Publishing
// and returns the first Delivery object with the matching correlationId.
// If the context times out while waiting for a reply, an error will be returned.
func DefaultDeliverer[Request, Response any](
	ctx context.Context,
	p Publisher[Request, Response],
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

// SendAndForgetDeliverer delivers the supplied publishing and
// returns a nil response.
// When using this deliverer please ensure that the supplied DecodeResponseFunc and
// PublisherResponseFunc are able to handle nil-type responses.
func SendAndForgetDeliverer[Request, Response any](
	ctx context.Context,
	p Publisher[Request, Response],
	pub *amqp.Publishing,
) (*amqp.Delivery, error) {
	err := p.ch.Publish(
		getPublishExchange(ctx),
		getPublishKey(ctx),
		false, //mandatory
		false, //immediate
		*pub,
	)
	return nil, err
}
