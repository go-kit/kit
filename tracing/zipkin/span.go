package zipkin

import (
	"errors"
	"time"

	"github.com/peterbourgon/gokit/tracing/zipkin/thrift/gen-go/zipkincore"
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
	host      string
	collector Collector

	name         string
	traceID      int64
	spanID       int64
	parentSpanID int64

	annotations []annotation
	//binaryAnnotations []BinaryAnnotation
}

// NewSpan returns a new Span object ready for use.
func NewSpan(host string, collector Collector, name string, traceID, spanID, parentSpanID int64) *Span {
	return &Span{
		host:         host,
		collector:    collector,
		name:         name,
		traceID:      traceID,
		spanID:       spanID,
		parentSpanID: parentSpanID,
	}
}

// NewSpanFunc returns a function that generates a new Zipkin span.
func NewSpanFunc(host string, collector Collector) func(string, int64, int64, int64) *Span {
	return func(name string, traceID, spanID, parentSpanID int64) *Span {
		return NewSpan(host, collector, name, traceID, spanID, parentSpanID)
	}
}

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

// Submit sends the span to the collector.
func (s *Span) Submit() error { return s.collector.Collect(s) }

// Encode creates a Thrift Span from the gokit Span.
func (s *Span) Encode() *zipkincore.Span {
	// TODO lots of garbage here. We can improve by preallocating e.g. the
	// Thrift stuff into an encoder struct, owned by the ScribeCollector.
	zs := zipkincore.Span{
		TraceId:           s.traceID,
		Name:              s.name,
		Id:                s.spanID,
		BinaryAnnotations: []*zipkincore.BinaryAnnotation{}, // TODO
		Debug:             false,                            // TODO
	}
	if s.parentSpanID != 0 {
		(*zs.ParentId) = s.parentSpanID
	}
	zs.Annotations = make([]*zipkincore.Annotation, len(s.annotations))
	for i, a := range s.annotations {
		zs.Annotations[i] = &zipkincore.Annotation{
			Timestamp: a.timestamp.UnixNano() / 1e3,
			Value:     a.value,
		}
		if a.host != "" {
			// zs.Annotations[i].Host = TODO
		}
		if a.duration > 0 {
			zs.Annotations[i].Duration = new(int32)
			(*zs.Annotations[i].Duration) = int32(a.duration / time.Microsecond)
		}
	}
	return &zs
}

type annotation struct {
	timestamp time.Time
	value     string
	duration  time.Duration // optional
	host      string
}
