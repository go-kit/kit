package amqp

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

// DecodeRequestFunc extracts a user-domain request object from
// an AMQP Delivery object. It is designed to be used in AMQP Subscribers.
type DecodeRequestFunc[Request any] func(context.Context, *amqp.Delivery) (request Request, err error)

// EncodeRequestFunc encodes the passed request object into
// an AMQP Publishing object. It is designed to be used in AMQP Publishers.
type EncodeRequestFunc[Request any] func(context.Context, *amqp.Publishing, Request) error

// EncodeResponseFunc encodes the passed response object to
// an AMQP Publishing object. It is designed to be used in AMQP Subscribers.
type EncodeResponseFunc[Response any] func(context.Context, *amqp.Publishing, Response) error

// DecodeResponseFunc extracts a user-domain response object from
// an AMQP Delivery object. It is designed to be used in AMQP Publishers.
type DecodeResponseFunc[Response any] func(context.Context, *amqp.Delivery) (response Response, err error)
