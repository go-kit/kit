package opentracing

import (
	"net"
	"net/http"
	"strconv"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"golang.org/x/net/context"

	"github.com/go-kit/kit/log"
	kithttp "github.com/go-kit/kit/transport/http"
)

// ToHTTPRequest returns an http RequestFunc that injects an OpenTracing Span
// found in `ctx` into the http headers. If no such Span can be found, the
// RequestFunc is a noop.
//
// The logger is used to report errors and may be nil.
func ToHTTPRequest(tracer opentracing.Tracer, logger log.Logger) kithttp.RequestFunc {
	return func(ctx context.Context, req *http.Request) context.Context {
		// Try to find a Span in the Context.
		if span := opentracing.SpanFromContext(ctx); span != nil {
			// Add standard OpenTracing tags.
			ext.HTTPMethod.Set(span, req.URL.RequestURI())
			host, portString, err := net.SplitHostPort(req.URL.Host)
			if err == nil {
				ext.PeerHostname.Set(span, host)
				if port, err := strconv.Atoi(portString); err != nil {
					ext.PeerPort.Set(span, uint16(port))
				}
			} else {
				ext.PeerHostname.Set(span, req.URL.Host)
			}

			// There's nothing we can do with any errors here.
			if err = tracer.Inject(
				span,
				opentracing.TextMap,
				opentracing.HTTPHeaderTextMapCarrier(req.Header),
			); err != nil {
				logger.Log("err", err)
			}
		}
		return ctx
	}
}

// FromHTTPRequest returns an http RequestFunc that tries to join with an
// OpenTracing trace found in `req` and starts a new Span called
// `operationName` accordingly. If no trace could be found in `req`, the Span
// will be a trace root. The Span is incorporated in the returned Context and
// can be retrieved with opentracing.SpanFromContext(ctx).
//
// The logger is used to report errors and may be nil.
func FromHTTPRequest(tracer opentracing.Tracer, operationName string, logger log.Logger) kithttp.RequestFunc {
	return func(ctx context.Context, req *http.Request) context.Context {
		// Try to join to a trace propagated in `req`.
		span, err := tracer.Join(
			operationName,
			opentracing.TextMap,
			opentracing.HTTPHeaderTextMapCarrier(req.Header),
		)
		if err != nil {
			span = tracer.StartSpan(operationName)
			if err != opentracing.ErrTraceNotFound {
				logger.Log("err", err)
			}
		}
		return opentracing.ContextWithSpan(ctx, span)
	}
}
