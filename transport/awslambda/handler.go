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
	s := &Handler{
		e:            e,
		dec:          dec,
		enc:          enc,
		logger:       log.NewNopLogger(),
		errorEncoder: DefaultErrorEncoder,
	}
	for _, option := range options {
		option(s)
	}
	return s
}

// HandlerOption sets an optional parameter for handlers.
type HandlerOption func(*Handler)

// HandlerBefore functions are executed on the payload byte,
// before the request is decoded.
func HandlerBefore(before ...HandlerRequestFunc) HandlerOption {
	return func(s *Handler) { s.before = append(s.before, before...) }
}

// HandlerAfter functions are only executed after invoking the endpoint
// but prior to returning a response.
func HandlerAfter(after ...HandlerResponseFunc) HandlerOption {
	return func(s *Handler) { s.after = append(s.after, after...) }
}

// HandlerErrorLogger is used to log non-terminal errors.
// By default, no errors are logged.
func HandlerErrorLogger(logger log.Logger) HandlerOption {
	return func(s *Handler) { s.logger = logger }
}

// HandlerErrorEncoder is used to encode errors.
func HandlerErrorEncoder(ee ErrorEncoder) HandlerOption {
	return func(s *Handler) { s.errorEncoder = ee }
}

// HandlerFinalizer sets finalizer which are called at the end of
// request. By default no finalizer is registered.
func HandlerFinalizer(f ...HandlerFinalizerFunc) HandlerOption {
	return func(s *Handler) { s.finalizer = append(s.finalizer, f...) }
}

// DefaultErrorEncoder defines the default behavior of encoding an error response,
// where it returns nil, and the error itself.
func DefaultErrorEncoder(ctx context.Context, err error) ([]byte, error) {
	return nil, err
}

// Invoke represents implementation of the AWS lambda.Handler interface.
func (s *Handler) Invoke(
	ctx context.Context,
	payload []byte,
) (resp []byte, err error) {
	if len(s.finalizer) > 0 {
		defer func() {
			for _, f := range s.finalizer {
				f(ctx, resp, err)
			}
		}()
	}

	for _, f := range s.before {
		ctx = f(ctx, payload)
	}

	request, err := s.dec(ctx, payload)
	if err != nil {
		s.logger.Log("err", err)
		resp, err = s.errorEncoder(ctx, err)
		return
	}

	response, err := s.e(ctx, request)
	if err != nil {
		s.logger.Log("err", err)
		resp, err = s.errorEncoder(ctx, err)
		return
	}

	for _, f := range s.after {
		ctx = f(ctx, response)
	}

	if resp, err = s.enc(ctx, response); err != nil {
		s.logger.Log("err", err)
		resp, err = s.errorEncoder(ctx, err)
		return
	}

	return resp, err
}
