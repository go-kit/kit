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

	// GetSpanDetails holds the function to use for generating the span name
	// based on the current name and from the information found in the incoming Request.
	// It can also return additional attributes for the span.
	//
	// A returned empty name defaults to the name that the middleware was initialized with.
	GetSpanDetails func(ctx context.Context, name string) (string, []trace.Attribute)
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

// WithSpanDetails extracts details from the request context (like span name and additional attributes).
func WithSpanDetails(fn func(ctx context.Context, name string) (string, []trace.Attribute)) EndpointOption {
	return func(o *EndpointOptions) {
		o.GetSpanDetails = fn
	}
}
