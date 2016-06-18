package pubsub

import "io"

// Publisher is a minimal interface for publishing messages to a pool of
// subscribers. Publishers are probably (but not necessarily) sending to a
// message bus.
//
// Most paramaterization of the publisher (topology restrictions like a topic,
// exchange, or specific message type; queue or buffer sizes; etc.) should be
// done in the concrete constructor.
type Publisher interface {
	// Publish a single message, described by an io.Reader, to the given key.
	Publish(key string, r io.Reader) error

	// Stop the publisher.
	Stop() error
}

// Subscriber is a minimal interface for subscribing to published messages.
// Subscribers are probably (but not necessarily) receiving from a message bus.
//
// Most paramaterization of the subscriber (topology restrictions like a topic,
// exchange, or specific message type; queue or buffer sizes; etc.) should be
// done in the concrete constructor.
type Subscriber interface {
	// Start returns a channel of messages that the caller should consume.
	// Failure to keep up with the incoming messages will have different
	// consequences depending on the concrete implementation of the subscriber.
	//
	// The channel will be closed when the subscriber encounters an error, or
	// when the caller invokes Stop, whichever comes first.
	Start() <-chan Message

	// Err returns the error that was responsible for closing the channel of
	// incoming messages.
	Err() error

	// Stop the subscriber, closing the channel that was returned by Start.
	Stop() error
}

// Message is a minimal interface to describe payloads received by subscribers.
// Clients may type-assert to more concrete types (e.g. pubsub/kafka.Message) to
// get access to more specific behaviors.
type Message interface {
	// Messages implement io.Reader to access the payload data.
	io.Reader

	// Done indicates the client is finished with the message, and the
	// underlying implementation may free its resources. Clients should ensure
	// to call Done for every received message.
	Done() error
}
