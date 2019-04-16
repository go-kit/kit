package log

// ErrorHandler is a transport error handler implementation which logs an error.
type ErrorHandler struct {
	logger Logger
}

func NewErrorHandler(logger Logger) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
	}
}

func (h *ErrorHandler) Handle(err error) {
	_ = h.logger.Log("err", err)
}
