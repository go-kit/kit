package opencensus

import (
	"context"

	"go.opencensus.io/trace"
)

// EndpointOptions holds the options for tracing an endpoint
type EndpointOptions struct {
	// IgnoreBusinessError if set to true will not treat a business error
	// identified through the endpoint.Failer interface as a span error.
	IgnoreBusinessError bool

	// Attributes holds the default attributes which will be set on span
	// creation by our Endpoint middleware.
	Attributes []trace.Attribute

	// GetName holds the function used for generating the span name
	// based on the current name and from the information found in the incoming Request.
	//
	// If the returned name is empty, the existing name for the endpoint is kept.
	GetName func(ctx context.Context, name string) string

	// GetAttributes holds the function used for extracting additional attributes
	// from the information found in the incoming Request.
	GetAttributes func(ctx context.Context) []trace.Attribute
}

// EndpointOption allows for functional options to our OpenCensus endpoint
// tracing middleware.
type EndpointOption func(*EndpointOptions)

// WithEndpointConfig sets all configuration options at once by use of the
// EndpointOptions struct.
func WithEndpointConfig(options EndpointOptions) EndpointOption {
	return func(o *EndpointOptions) {
		*o = options
	}
}

// WithEndpointAttributes sets the default attributes for the spans created by
// the Endpoint tracer.
func WithEndpointAttributes(attrs ...trace.Attribute) EndpointOption {
	return func(o *EndpointOptions) {
		o.Attributes = attrs
	}
}

// WithIgnoreBusinessError if set to true will not treat a business error
// identified through the endpoint.Failer interface as a span error.
func WithIgnoreBusinessError(val bool) EndpointOption {
	return func(o *EndpointOptions) {
		o.IgnoreBusinessError = val
	}
}

// WithSpanName extracts additional attributes from the request context.
func WithSpanName(fn func(ctx context.Context, name string) string) EndpointOption {
	return func(o *EndpointOptions) {
		o.GetName = fn
	}
}

// WithSpanAttributes extracts additional attributes from the request context.
func WithSpanAttributes(fn func(ctx context.Context) []trace.Attribute) EndpointOption {
	return func(o *EndpointOptions) {
		o.GetAttributes = fn
	}
}
