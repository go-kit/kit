package zipkin

import (
	"errors"
	"time"
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

type annotation struct {
	timestamp time.Time
	value     string
	duration  time.Duration // optional
	host      string
}
