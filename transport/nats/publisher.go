package nats

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/openmesh/kit/endpoint"
)

// Publisher wraps a URL and provides a method that implements endpoint.Endpoint.
type Publisher[Request, Response any] struct {
	publisher *nats.Conn
	subject   string
	enc       EncodeRequestFunc[Request]
	dec       DecodeResponseFunc[Response]
	before    []RequestFunc
	after     []PublisherResponseFunc
	timeout   time.Duration
}

// NewPublisher constructs a usable Publisher for a single remote method.
func NewPublisher[Request, Response any](
	publisher *nats.Conn,
	subject string,
	enc EncodeRequestFunc[Request],
	dec DecodeResponseFunc[Response],
	options ...PublisherOption[Request, Response],
) *Publisher[Request, Response] {
	p := &Publisher[Request, Response]{
		publisher: publisher,
		subject:   subject,
		enc:       enc,
		dec:       dec,
		timeout:   10 * time.Second,
	}
	for _, option := range options {
		option(p)
	}
	return p
}

// PublisherOption sets an optional parameter for clients.
type PublisherOption[Request, Response any] func(*Publisher[Request, Response])

// PublisherBefore sets the RequestFuncs that are applied to the outgoing NATS
// request before it's invoked.
func PublisherBefore[Request, Response any](before ...RequestFunc) PublisherOption[Request, Response] {
	return func(p *Publisher[Request, Response]) { p.before = append(p.before, before...) }
}

// PublisherAfter sets the ClientResponseFuncs applied to the incoming NATS
// request prior to it being decoded. This is useful for obtaining anything off
// of the response and adding onto the context prior to decoding.
func PublisherAfter[Request, Response any](after ...PublisherResponseFunc) PublisherOption[Request, Response] {
	return func(p *Publisher[Request, Response]) { p.after = append(p.after, after...) }
}

// PublisherTimeout sets the available timeout for NATS request.
func PublisherTimeout[Request, Response any](timeout time.Duration) PublisherOption[Request, Response] {
	return func(p *Publisher[Request, Response]) { p.timeout = timeout }
}

// Endpoint returns a usable endpoint that invokes the remote endpoint.
func (p Publisher[Request, Response]) Endpoint() endpoint.Endpoint[Request, Response] {
	return func(ctx context.Context, request Request) (Response, error) {
		ctx, cancel := context.WithTimeout(ctx, p.timeout)
		defer cancel()

		msg := nats.Msg{Subject: p.subject}

		if err := p.enc(ctx, &msg, request); err != nil {
			return *new(Response), err
		}

		for _, f := range p.before {
			ctx = f(ctx, &msg)
		}

		resp, err := p.publisher.RequestWithContext(ctx, msg.Subject, msg.Data)
		if err != nil {
			return *new(Response), err
		}

		for _, f := range p.after {
			ctx = f(ctx, resp)
		}

		response, err := p.dec(ctx, resp)
		if err != nil {
			return *new(Response), err
		}

		return response, nil
	}
}

// EncodeJSONRequest is an EncodeRequestFunc that serializes the request as a
// JSON object to the Data of the Msg. Many JSON-over-NATS services can use it as
// a sensible default.
func EncodeJSONRequest(_ context.Context, msg *nats.Msg, request interface{}) error {
	b, err := json.Marshal(request)
	if err != nil {
		return err
	}

	msg.Data = b

	return nil
}
