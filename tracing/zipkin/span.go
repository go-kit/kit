package zipkin

import (
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"strconv"
	"time"

	"github.com/go-kit/kit/tracing/zipkin/_thrift/gen-go/zipkincore"
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

	annotations       []annotation
	binaryAnnotations []binaryAnnotation
}

// NewSpan returns a new Span, which can be annotated and collected by a
// collector. Spans are passed through the request context to each middleware
// under the SpanContextKey.
func NewSpan(hostport, serviceName, methodName string, traceID, spanID, parentSpanID int64) *Span {
	return &Span{
		host:         makeEndpoint(hostport, serviceName),
		methodName:   methodName,
		traceID:      traceID,
		spanID:       spanID,
		parentSpanID: parentSpanID,
	}
}

// makeEndpoint takes the hostport and service name that represent this Zipkin
// service, and returns an endpoint that's embedded into the Zipkin core Span
// type. It will return a nil endpoint if the input parameters are malformed.
func makeEndpoint(hostport, serviceName string) *zipkincore.Endpoint {
	host, port, err := net.SplitHostPort(hostport)
	if err != nil {
		return nil
	}
	portInt, err := strconv.ParseInt(port, 10, 16)
	if err != nil {
		return nil
	}
	addrs, err := net.LookupIP(host)
	if err != nil {
		return nil
	}
	// we need the first IPv4 address.
	var addr net.IP
	for i := range addrs {
		addr = addrs[i].To4()
		if addr != nil {
			break
		}
	}
	if addr == nil {
		// none of the returned addresses is IPv4.
		return nil
	}
	endpoint := zipkincore.NewEndpoint()
	endpoint.Ipv4 = (int32)(binary.BigEndian.Uint32(addr))
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

// AnnotateBinary annotates the span with a key and a value that will be []byte
// encoded.
func (s *Span) AnnotateBinary(key string, value interface{}) {
	var a zipkincore.AnnotationType
	var b []byte
	// We are not using zipkincore.AnnotationType_I16 for types that could fit
	// as reporting on it seems to be broken on the zipkin web interface
	// (however, we can properly extract the number from zipkin storage
	// directly). int64 has issues with negative numbers but seems ok for
	// positive numbers needing more than 32 bit.
	switch v := value.(type) {
	case bool:
		a = zipkincore.AnnotationType_BOOL
		b = []byte("\x00")
		if v {
			b = []byte("\x01")
		}
	case []byte:
		a = zipkincore.AnnotationType_BYTES
		b = v
	case byte:
		a = zipkincore.AnnotationType_I32
		b = make([]byte, 4)
		binary.BigEndian.PutUint32(b, uint32(v))
	case int8:
		a = zipkincore.AnnotationType_I32
		b = make([]byte, 4)
		binary.BigEndian.PutUint32(b, uint32(v))
	case int16:
		a = zipkincore.AnnotationType_I32
		b = make([]byte, 4)
		binary.BigEndian.PutUint32(b, uint32(v))
	case uint16:
		a = zipkincore.AnnotationType_I32
		b = make([]byte, 4)
		binary.BigEndian.PutUint32(b, uint32(v))
	case int32:
		a = zipkincore.AnnotationType_I32
		b = make([]byte, 4)
		binary.BigEndian.PutUint32(b, uint32(v))
	case uint32:
		a = zipkincore.AnnotationType_I32
		b = make([]byte, 4)
		binary.BigEndian.PutUint32(b, uint32(v))
	case int64:
		a = zipkincore.AnnotationType_I64
		b = make([]byte, 8)
		binary.BigEndian.PutUint64(b, uint64(v))
	case int:
		a = zipkincore.AnnotationType_I32
		b = make([]byte, 8)
		binary.BigEndian.PutUint32(b, uint32(v))
	case uint:
		a = zipkincore.AnnotationType_I32
		b = make([]byte, 8)
		binary.BigEndian.PutUint32(b, uint32(v))
	case uint64:
		a = zipkincore.AnnotationType_I64
		b = make([]byte, 8)
		binary.BigEndian.PutUint64(b, uint64(v))
	case float32:
		a = zipkincore.AnnotationType_DOUBLE
		b = make([]byte, 8)
		bits := math.Float64bits(float64(v))
		binary.BigEndian.PutUint64(b, bits)
	case float64:
		a = zipkincore.AnnotationType_DOUBLE
		b = make([]byte, 8)
		bits := math.Float64bits(v)
		binary.BigEndian.PutUint64(b, bits)
	case string:
		a = zipkincore.AnnotationType_STRING
		b = []byte(v)
	default:
		// we have no handler for type's value, but let's get a string
		// representation of it.
		a = zipkincore.AnnotationType_STRING
		b = []byte(fmt.Sprintf("%+v", value))
	}
	s.binaryAnnotations = append(s.binaryAnnotations, binaryAnnotation{
		key:            key,
		value:          b,
		annotationType: a,
		host:           s.host,
	})
}

// AnnotateString annotates the span with a key and a string value.
func (s *Span) AnnotateString(key, value string) {
	s.binaryAnnotations = append(s.binaryAnnotations, binaryAnnotation{
		key:            key,
		value:          []byte(value),
		annotationType: zipkincore.AnnotationType_STRING,
		host:           s.host,
	})
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
		TraceId: s.traceID,
		Name:    s.methodName,
		Id:      s.spanID,
		Debug:   true, // TODO
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

	zs.BinaryAnnotations = make([]*zipkincore.BinaryAnnotation, len(s.binaryAnnotations))
	for i, a := range s.binaryAnnotations {
		zs.BinaryAnnotations[i] = &zipkincore.BinaryAnnotation{
			Key:            a.key,
			Value:          a.value,
			AnnotationType: a.annotationType,
			Host:           a.host,
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

type binaryAnnotation struct {
	key            string
	value          []byte
	annotationType zipkincore.AnnotationType
	host           *zipkincore.Endpoint
}
