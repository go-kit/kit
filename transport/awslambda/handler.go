package awslambda

import (
	"context"

	"github.com/go-kit/log"
	"github.com/openmesh/kit/endpoint"
	"github.com/openmesh/kit/transport"
)

// Handler wraps an endpoint.
type Handler[Request, Response any] struct {
	e            endpoint.Endpoint[Request, Response]
	dec          DecodeRequestFunc[Request]
	enc          EncodeResponseFunc[Response]
	before       []HandlerRequestFunc
	after        []HandlerResponseFunc
	errorEncoder ErrorEncoder
	finalizer    []HandlerFinalizerFunc
	errorHandler transport.ErrorHandler
}

// NewHandler constructs a new handler, which implements
// the AWS lambda.Handler interface.
func NewHandler[Request, Response any](
	e endpoint.Endpoint[Request, Response],
	dec DecodeRequestFunc[Request],
	enc EncodeResponseFunc[Response],
	options ...HandlerOption[Request, Response],
) *Handler[Request, Response] {
	h := &Handler[Request, Response]{
		e:            e,
		dec:          dec,
		enc:          enc,
		errorEncoder: DefaultErrorEncoder,
		errorHandler: transport.NewLogErrorHandler(log.NewNopLogger()),
	}
	for _, option := range options {
		option(h)
	}
	return h
}

// HandlerOption sets an optional parameter for handlers.
type HandlerOption[Request, Response any] func(*Handler[Request, Response])

// HandlerBefore functions are executed on the payload byte,
// before the request is decoded.
func HandlerBefore[Request, Response any](before ...HandlerRequestFunc) HandlerOption[Request, Response] {
	return func(h *Handler[Request, Response]) { h.before = append(h.before, before...) }
}

// HandlerAfter functions are only executed after invoking the endpoint
// but prior to returning a response.
func HandlerAfter[Request, Response any](after ...HandlerResponseFunc) HandlerOption[Request, Response] {
	return func(h *Handler[Request, Response]) { h.after = append(h.after, after...) }
}

// HandlerErrorLogger is used to log non-terminal errors.
// By default, no errors are logged.
// Deprecated: Use HandlerErrorHandler instead.
func HandlerErrorLogger[Request, Response any](logger log.Logger) HandlerOption[Request, Response] {
	return func(h *Handler[Request, Response]) { h.errorHandler = transport.NewLogErrorHandler(logger) }
}

// HandlerErrorHandler is used to handle non-terminal errors.
// By default, non-terminal errors are ignored.
func HandlerErrorHandler[Request, Response any](errorHandler transport.ErrorHandler) HandlerOption[Request, Response] {
	return func(h *Handler[Request, Response]) { h.errorHandler = errorHandler }
}

// HandlerErrorEncoder is used to encode errors.
func HandlerErrorEncoder[Request, Response any](ee ErrorEncoder) HandlerOption[Request, Response] {
	return func(h *Handler[Request, Response]) { h.errorEncoder = ee }
}

// HandlerFinalizer sets finalizer which are called at the end of
// request. By default no finalizer is registered.
func HandlerFinalizer[Request, Response any](f ...HandlerFinalizerFunc) HandlerOption[Request, Response] {
	return func(h *Handler[Request, Response]) { h.finalizer = append(h.finalizer, f...) }
}

// DefaultErrorEncoder defines the default behavior of encoding an error response,
// where it returns nil, and the error itself.
func DefaultErrorEncoder(ctx context.Context, err error) ([]byte, error) {
	return nil, err
}

// Invoke represents implementation of the AWS lambda.Handler interface.
func (h *Handler[Request, Response]) Invoke(
	ctx context.Context,
	payload []byte,
) (resp []byte, err error) {
	if len(h.finalizer) > 0 {
		defer func() {
			for _, f := range h.finalizer {
				f(ctx, resp, err)
			}
		}()
	}

	for _, f := range h.before {
		ctx = f(ctx, payload)
	}

	request, err := h.dec(ctx, payload)
	if err != nil {
		h.errorHandler.Handle(ctx, err)
		return h.errorEncoder(ctx, err)
	}

	response, err := h.e(ctx, request)
	if err != nil {
		h.errorHandler.Handle(ctx, err)
		return h.errorEncoder(ctx, err)
	}

	for _, f := range h.after {
		ctx = f(ctx, response)
	}

	if resp, err = h.enc(ctx, response); err != nil {
		h.errorHandler.Handle(ctx, err)
		return h.errorEncoder(ctx, err)
	}

	return resp, err
}
