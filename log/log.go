package log

import "errors"

// Logger is the fundamental interface for all log operations. Log creates a
// log event from keyvals, a variadic sequence of alternating keys and values.
// Implementations must be safe for concurrent use by multiple goroutines. In
// particular, any implementation of Logger that appends to keyvals or
// modifies any of its elements must make a copy first.
type Logger interface {
	Log(keyvals ...interface{}) error
}

// ErrMissingValue is appended to keyvals slices with odd length to substitute
// the missing value.
var ErrMissingValue = errors.New("(MISSING)")

// NewContext returns a new Context that logs to logger.
func NewContext(logger Logger) *Context {
	if c, ok := logger.(*Context); ok {
		return c
	}
	return &Context{logger: logger}
}

// Context must always have the same number of stack frames between calls to
// its Log method and the eventual binding of Valuers to their value. This
// requirement comes from the functional requirement to allow a context to
// resolve application call site information for a log.Caller stored in the
// context. To do this we must be able to predict the number of logging
// functions on the stack when bindValues is called.
//
// Three implementation details provide the needed stack depth consistency.
// The first two of these details also result in better amortized performance,
// and thus make sense even without the requirements regarding stack depth.
// The third detail, however, is subtle and tied to the implementation of the
// Go compiler.
//
//    1. NewContext avoids introducing an additional layer when asked to
//       wrap another Context.
//    2. With avoids introducing an additional layer by returning a newly
//       constructed Context with a merged keyvals rather than simply
//       wrapping the existing Context.
//    3. All of Context's methods take pointer receivers even though they
//       do not mutate the Context.
//
// Before explaining the last detail, first some background. The Go compiler
// generates wrapper methods to implement the auto dereferencing behavior when
// calling a value method through a pointer variable. These wrapper methods
// are also used when calling a value method through an interface variable
// because interfaces store a pointer to the underlying concrete value.
// Calling a pointer receiver through an interface does not require generating
// an additional function.
//
// If Context had value methods then calling Context.Log through a variable
// with type Logger would have an extra stack frame compared to calling
// Context.Log through a variable with type Context. Using pointer receivers
// avoids this problem.

// A Context wraps a Logger and holds keyvals that it includes in all log
// events. When logging, a Context replaces all value elements (odd indexes)
// containing a Valuer with their generated value for each call to its Log
// method.
type Context struct {
	logger     Logger
	keyvals    []interface{}
	hasValuer  bool
	projection *deferrableProjection
}

// Log replaces all value elements (odd indexes) containing a Valuer in the
// stored context with their generated value, appends keyvals, and passes the
// result to the wrapped Logger.
func (l *Context) Log(keyvals ...interface{}) error {
	if l == nil {
		return nil
	}
	applyDeferredProjection := false
	if l.projection != nil {
		if l.projection.deferred {
			applyDeferredProjection = true
		} else {
			if len(keyvals)%2 != 0 {
				keyvals = append(keyvals, ErrMissingValue)
			}
			projected, preserve := l.projection.Projection(keyvals)
			if !preserve {
				return nil
			}
			keyvals = projected
		}
	}
	kvs := append(l.keyvals, keyvals...)
	if len(kvs)%2 != 0 {
		kvs = append(kvs, ErrMissingValue)
	}
	if l.hasValuer {
		// If no keyvals were appended above then we must copy l.keyvals so
		// that future log events will reevaluate the stored Valuers.
		if len(keyvals) == 0 {
			kvs = append([]interface{}{}, l.keyvals...)
		}
		bindValues(kvs[:len(l.keyvals)])
	}
	if applyDeferredProjection {
		projected, preserve := l.projection.Projection(kvs)
		if !preserve {
			return nil
		}
		if len(projected)%2 != 0 {
			kvs = append(projected, ErrMissingValue)
		} else {
			kvs = projected
		}
	}
	return l.logger.Log(kvs...)
}

// With returns a new Context with keyvals appended to those of the receiver.
func (l *Context) With(keyvals ...interface{}) *Context {
	if len(keyvals) == 0 || l == nil {
		return l
	}
	if l.projection != nil && !l.projection.deferred {
		if len(keyvals)%2 != 0 {
			keyvals = append(keyvals, ErrMissingValue)
		}
		projected, preserve := l.projection.Projection(keyvals)
		if !preserve {
			return nil
		}
		keyvals = projected
	}
	kvs := append(l.keyvals, keyvals...)
	if len(kvs)%2 != 0 {
		kvs = append(kvs, ErrMissingValue)
	}
	return &Context{
		logger: l.logger,
		// Limiting the capacity of the stored keyvals ensures that a new
		// backing array is created if the slice must grow in Log or With.
		// Using the extra capacity without copying risks a data race that
		// would violate the Logger interface contract.
		keyvals:    kvs[:len(kvs):len(kvs)],
		hasValuer:  l.hasValuer || containsValuer(keyvals),
		projection: l.projection,
	}
}

// WithPrefix returns a new Context with keyvals prepended to those of the
// receiver.
func (l *Context) WithPrefix(keyvals ...interface{}) *Context {
	if len(keyvals) == 0 || l == nil {
		return l
	}
	if l.projection != nil && !l.projection.deferred {
		if len(keyvals)%2 != 0 {
			keyvals = append(keyvals, ErrMissingValue)
		}
		projected, preserve := l.projection.Projection(keyvals)
		if !preserve {
			return nil
		}
		keyvals = projected
	}
	// Limiting the capacity of the stored keyvals ensures that a new
	// backing array is created if the slice must grow in Log or With.
	// Using the extra capacity without copying risks a data race that
	// would violate the Logger interface contract.
	n := len(l.keyvals) + len(keyvals)
	if len(keyvals)%2 != 0 {
		n++
	}
	kvs := make([]interface{}, 0, n)
	kvs = append(kvs, keyvals...)
	if len(kvs)%2 != 0 {
		kvs = append(kvs, ErrMissingValue)
	}
	kvs = append(kvs, l.keyvals...)
	return &Context{
		logger:     l.logger,
		keyvals:    kvs,
		hasValuer:  l.hasValuer || containsValuer(keyvals),
		projection: l.projection,
	}
}

// A Projection inspects a proposed set of keys and values for a log
// record, deciding whether to preserve or drop the record and
// optionally proposing a different set of keys and values to use
// instead.
//
// It must not modify the supplied slice, and should return an even
// number of keys and values paired in an alternating sequence.
type Projection func(keyvals []interface{}) (kvs []interface{}, preserve bool)

type deferrableProjection struct {
	Projection
	deferred bool
}

// WithPreProjection returns a new Context that applies the supplied
// projection function before any other such functions already bound
// to the original Context.
//
// If deferred, it waits to apply the function only to complete log
// records before submitting them first to any other projection
// functions and then to its downstream logger. Otherwise, it applies
// the function to keys and values as they accumulate in calls to
// With, WithPrefix, and Log, before interpolating any contributions
// from Valuers.
func (l *Context) WithPreProjection(p Projection, deferred bool) *Context {
	if p == nil || l == nil {
		return l
	}
	if outer := l.projection; outer != nil {
		if !deferred && outer.deferred {
			deferred = true
		}
		inner := p
		p = func(keyvals []interface{}) ([]interface{}, bool) {
			kvs, preserve := inner(keyvals)
			if !preserve {
				return nil, false
			}
			return outer.Projection(kvs)
		}
	}
	return &Context{
		logger:     l.logger,
		keyvals:    l.keyvals,
		hasValuer:  l.hasValuer,
		projection: &deferrableProjection{p, deferred},
	}
}

// WithPostProjection returns a new Context that applies the supplied
// projection function after any other such functions already bound
// to the original Context.
//
// If deferred, it waits to apply the function only to complete log
// records as returned from any other projection functions, and
// submits the result to its downstream logger. Otherwise, it applies
// the function to keys and values as they accumulate in calls to
// With, WithPrefix, and Log, before interpolating any contributions
// from Valuers.
func (l *Context) WithPostProjection(p Projection, deferred bool) *Context {
	if p == nil || l == nil {
		return l
	}
	if inner := l.projection; inner != nil {
		if !deferred && inner.deferred {
			deferred = true
		}
		outer := p
		p = func(keyvals []interface{}) ([]interface{}, bool) {
			kvs, preserve := inner.Projection(keyvals)
			if !preserve {
				return nil, false
			}
			return outer(kvs)
		}
	}
	return &Context{
		logger:     l.logger,
		keyvals:    l.keyvals,
		hasValuer:  l.hasValuer,
		projection: &deferrableProjection{p, deferred},
	}
}

// LoggerFunc is an adapter to allow use of ordinary functions as Loggers. If
// f is a function with the appropriate signature, LoggerFunc(f) is a Logger
// object that calls f.
type LoggerFunc func(...interface{}) error

// Log implements Logger by calling f(keyvals...).
func (f LoggerFunc) Log(keyvals ...interface{}) error {
	return f(keyvals...)
}
