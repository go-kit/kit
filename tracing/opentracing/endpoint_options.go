package opentracing

import (
	"context"
	"github.com/opentracing/opentracing-go"
)

// EndpointOptions holds the options for tracing an endpoint
type EndpointOptions struct {
	// IgnoreBusinessError if set to true will not treat a business error
	// identified through the endpoint.Failer interface as a span error.
	IgnoreBusinessError bool

	// GetOperationName is an optional function that can set the span operation name based on the existing one
	// for the endpoint and information in the context.
	//
	// If the function is nil, or the returned name is empty, the existing name for the endpoint is used.
	GetOperationName func(ctx context.Context, name string) string

	// Tags holds the default tags which will be set on span
	// creation by our Endpoint middleware.
	Tags opentracing.Tags

	// GetTags is an optional function that can extract tags
	// from the context and add them to the span.
	GetTags func(ctx context.Context) opentracing.Tags
}

// EndpointOption allows for functional options to endpoint tracing middleware.
type EndpointOption func(*EndpointOptions)

// WithOptions sets all configuration options at once by use of the EndpointOptions struct.
func WithOptions(options EndpointOptions) EndpointOption {
	return func(o *EndpointOptions) {
		*o = options
	}
}

// WithIgnoreBusinessError if set to true will not treat a business error
// identified through the endpoint.Failer interface as a span error.
func WithIgnoreBusinessError(ignoreBusinessError bool) EndpointOption {
	return func(o *EndpointOptions) {
		o.IgnoreBusinessError = ignoreBusinessError
	}
}

// WithOperationName allows to set function that can set the span operation name based on the existing one
// for the endpoint and information in the context.
func WithOperationName(getOperationName func(ctx context.Context, name string) string) EndpointOption {
	return func(o *EndpointOptions) {
		o.GetOperationName = getOperationName
	}
}

// WithTags sets the default tags for the spans created by the Endpoint tracer.
func WithTags(tags opentracing.Tags) EndpointOption {
	return func(o *EndpointOptions) {
		o.Tags = tags
	}
}

// WithExtraTags extracts additional tags from the context.
func WithExtraTags(getTags func(ctx context.Context) opentracing.Tags) EndpointOption {
	return func(o *EndpointOptions) {
		o.GetTags = getTags
	}
}
