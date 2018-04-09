package flags

import "context"

// Booler describes a feature flag that returns a simple boolean response
type Booler interface {
	Bool(c context.Context) bool
}

// BoolerFunc is an adapter to use a stand-alone function as a Booler
type BoolerFunc func(c context.Context) bool

// Bool conforms to the Booler interface
func (fn BoolerFunc) Bool(c context.Context) bool {
	return fn(c)
}

// Inter describes a feature flag that returns a simple int64 response
type Inter interface {
	Int(c context.Context) int64
}

// InterFunc is an adapter to use a stand-alone function as an Inter
type InterFunc func(c context.Context) int64

// Int conforms to the Inter interface
func (fn InterFunc) Int(c context.Context) int64 {
	return fn(c)
}

// Floater describes a feature flag that returns a simple float64 response
type Floater interface {
	Float(c context.Context) float64
}

// FloaterFunc is an adapter to use a stand-alone function as an Floater
type FloaterFunc func(c context.Context) float64

// Float conforms to the Floater interface
func (fn FloaterFunc) Float(c context.Context) float64 {
	return fn(c)
}

// Stringer describes a feature flag that returns a simple string response
type Stringer interface {
	String(c context.Context) string
}

// StringerFunc is an adapter to use a stand-alone function as an Stringer
type StringerFunc func(c context.Context) string

// String conforms to the Stringer interface
func (fn StringerFunc) String(c context.Context) string {
	return fn(c)
}
