package amqp

// ErrorHandler receives a transport error to be processed for diagnostic purposes.
// Usually this means logging the error.
type ErrorHandler interface {
	Handle(err error)
}
