package zipkin

import (
	"encoding/binary"
	"errors"
	"net"
	"strconv"
	"time"

	"github.com/go-kit/kit/log"

	"github.com/go-kit/kit/tracing/zipkin/_thrift/gen-go/zipkincore"
)

var (
	// SpanContextKey represents the Span in the request context.
	SpanContextKey = "Zipkin-Span"

	// ErrSpanNotFound is returned when a Span isn't found in a context.
	ErrSpanNotFound = errors.New("span not found")
)

// A Span is a named collection of annotations. It represents meaningful
// information about a single method call, i.e. a single request against a
// service. Clients should annotate the span, and submit it when the request
// that generated it is complete.
type Span struct {
	host       *zipkincore.Endpoint
	methodName string

	traceID      int64
	spanID       int64
	parentSpanID int64

	annotations []annotation
	//binaryAnnotations []BinaryAnnotation
}

// NewSpan returns a new Span object ready for use.
func NewSpan(hostport, serviceName, methodName string, traceID, spanID, parentSpanID int64) *Span {
	return &Span{
		host:         makeEndpoint(hostport, serviceName),
		methodName:   methodName,
		traceID:      traceID,
		spanID:       spanID,
		parentSpanID: parentSpanID,
	}
}

// makeEndpoint will return a nil Endpoint if the input parameters are
// malformed.
func makeEndpoint(hostport, serviceName string) *zipkincore.Endpoint {
	host, port, err := net.SplitHostPort(hostport)
	if err != nil {
		log.DefaultLogger.Log("hostport", hostport, "err", err)
		return nil
	}
	addrs, err := net.LookupIP(host)
	if err != nil {
		log.DefaultLogger.Log("host", host, "err", err)
		return nil
	}
	if len(addrs) <= 0 {
		log.DefaultLogger.Log("host", host, "err", "no IPs")
		return nil
	}
	portInt, err := strconv.ParseInt(port, 10, 16)
	if err != nil {
		log.DefaultLogger.Log("port", port, "err", err)
		return nil
	}
	endpoint := zipkincore.NewEndpoint()
	binary.LittleEndian.PutUint32(addrs[0], (uint32)(endpoint.Ipv4))
	endpoint.Port = int16(portInt)
	endpoint.ServiceName = serviceName
	return endpoint
}

// MakeNewSpanFunc returns a function that generates a new Zipkin span.
func MakeNewSpanFunc(hostport, serviceName, methodName string) NewSpanFunc {
	return func(traceID, spanID, parentSpanID int64) *Span {
		return NewSpan(hostport, serviceName, methodName, traceID, spanID, parentSpanID)
	}
}

// NewSpanFunc takes trace, span, & parent span IDs to produce a Span object.
type NewSpanFunc func(traceID, spanID, parentSpanID int64) *Span

// TraceID returns the ID of the trace that this span is a member of.
func (s *Span) TraceID() int64 { return s.traceID }

// SpanID returns the ID of this span.
func (s *Span) SpanID() int64 { return s.spanID }

// ParentSpanID returns the ID of the span which invoked this span.
// It may be zero.
func (s *Span) ParentSpanID() int64 { return s.parentSpanID }

// Annotate annotates the span with the given value.
func (s *Span) Annotate(value string) {
	s.AnnotateDuration(value, 0)
}

// AnnotateDuration annotates the span with the given value and duration.
func (s *Span) AnnotateDuration(value string, duration time.Duration) {
	s.annotations = append(s.annotations, annotation{
		timestamp: time.Now(),
		value:     value,
		duration:  duration,
		host:      s.host,
	})
}

// Encode creates a Thrift Span from the gokit Span.
func (s *Span) Encode() *zipkincore.Span {
	// TODO lots of garbage here. We can improve by preallocating e.g. the
	// Thrift stuff into an encoder struct, owned by the ScribeCollector.
	zs := zipkincore.Span{
		TraceId:           s.traceID,
		Name:              s.methodName,
		Id:                s.spanID,
		BinaryAnnotations: []*zipkincore.BinaryAnnotation{}, // TODO
		Debug:             true,                             // TODO
	}
	if s.parentSpanID != 0 {
		zs.ParentId = new(int64)
		(*zs.ParentId) = s.parentSpanID
	}
	zs.Annotations = make([]*zipkincore.Annotation, len(s.annotations))
	for i, a := range s.annotations {
		zs.Annotations[i] = &zipkincore.Annotation{
			Timestamp: a.timestamp.UnixNano() / 1e3,
			Value:     a.value,
			Host:      a.host,
		}
		if a.duration > 0 {
			zs.Annotations[i].Duration = new(int32)
			*(zs.Annotations[i].Duration) = int32(a.duration / time.Microsecond)
		}
	}
	return &zs
}

type annotation struct {
	timestamp time.Time
	value     string
	duration  time.Duration // optional
	host      *zipkincore.Endpoint
}
