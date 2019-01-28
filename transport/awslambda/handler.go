package awslambda

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

// Handler wraps an endpoint.
type Handler struct {
	e            endpoint.Endpoint
	dec          DecodeRequestFunc
	enc          EncodeResponseFunc
	before       []HandlerRequestFunc
	after        []HandlerResponseFunc
	errorEncoder ErrorEncoder
	finalizer    []HandlerFinalizerFunc
	logger       log.Logger
}

// NewHandler constructs a new handler, which implements
// the AWS lambda.Handler interface.
func NewHandler(
	e endpoint.Endpoint,
	dec DecodeRequestFunc,
	enc EncodeResponseFunc,
	options ...HandlerOption,
) *Handler {
	h := &Handler{
		e:            e,
		dec:          dec,
		enc:          enc,
		logger:       log.NewNopLogger(),
		errorEncoder: DefaultErrorEncoder,
	}
	for _, option := range options {
		option(h)
	}
	return h
}

// HandlerOption sets an optional parameter for handlerh.
type HandlerOption func(*Handler)

// HandlerBefore functions are executed on the payload byte,
// before the request is decoded.
func HandlerBefore(before ...HandlerRequestFunc) HandlerOption {
	return func(h *Handler) { h.before = append(h.before, before...) }
}

// HandlerAfter functions are only executed after invoking the endpoint
// but prior to returning a response.
func HandlerAfter(after ...HandlerResponseFunc) HandlerOption {
	return func(h *Handler) { h.after = append(h.after, after...) }
}

// HandlerErrorLogger is used to log non-terminal errorh.
// By default, no errors are logged.
func HandlerErrorLogger(logger log.Logger) HandlerOption {
	return func(h *Handler) { h.logger = logger }
}

// HandlerErrorEncoder is used to encode errorh.
func HandlerErrorEncoder(ee ErrorEncoder) HandlerOption {
	return func(h *Handler) { h.errorEncoder = ee }
}

// HandlerFinalizer sets finalizer which are called at the end of
// request. By default no finalizer is registered.
func HandlerFinalizer(f ...HandlerFinalizerFunc) HandlerOption {
	return func(h *Handler) { h.finalizer = append(h.finalizer, f...) }
}

// DefaultErrorEncoder defines the default behavior of encoding an error response,
// where it returns nil, and the error itself.
func DefaultErrorEncoder(ctx context.Context, err error) ([]byte, error) {
	return nil, err
}

// Invoke represents implementation of the AWS lambda.Handler interface.
func (h *Handler) Invoke(
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
		h.logger.Log("err", err)
		resp, err = h.errorEncoder(ctx, err)
		return
	}

	response, err := h.e(ctx, request)
	if err != nil {
		h.logger.Log("err", err)
		resp, err = h.errorEncoder(ctx, err)
		return
	}

	for _, f := range h.after {
		ctx = f(ctx, response)
	}

	if resp, err = h.enc(ctx, response); err != nil {
		h.logger.Log("err", err)
		resp, err = h.errorEncoder(ctx, err)
		return
	}

	return resp, err
}
