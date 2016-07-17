// Package sterrors provides a structured error.
//
// You can return a structured error and log the error with key value pairs
// attached to it later at a higher call stack. This is especially useful
// for libraries since library does not need to log errors and library
// users can log errors in their application using their favorite logging
// package.
package sterrors

// ErrKeyValser is an interface for an error with key value pairs.
type ErrKeyValser interface {
	// Err returns the original error.
	Err() error

	// Keyvals returns keys and values.
	KeyVals() []interface{}
}

// With returns a new error with keyvals appended to those of err
// if err is sterr, or a new error with keyvals if not.
//
// It panics if the length of keyvals are an even number.
func With(err error, keyvals ...interface{}) error {
	if len(keyvals)%2 != 0 {
		panic("keyvals length must be even number")
	}
	if err2, ok := err.(ErrKeyValser); ok {
		return errWithKeyVals{
			err2.Err(),
			append(err2.KeyVals(), keyvals...),
		}
	} else {
		return errWithKeyVals{
			err,
			keyvals,
		}
	}
}

// With returns a new error with keyvals prepended to those of err
// if err is sterr, or a new error with keyvals if not.
//
// It panics if the length of keyvals are an even number.
func WithPrefix(err error, keyvals ...interface{}) error {
	if len(keyvals)%2 != 0 {
		panic("keyvals length must be even number")
	}
	if err2, ok := err.(ErrKeyValser); ok {
		return errWithKeyVals{
			err2.Err(),
			append(keyvals, err2.KeyVals()...),
		}
	} else {
		return errWithKeyVals{
			err,
			keyvals,
		}
	}
}

type errWithKeyVals struct {
	error
	keyvals []interface{}
}

// Err returns the original error.
func (e errWithKeyVals) Err() error {
	return e.error
}

// Keyvals returns keys and values.
func (e errWithKeyVals) KeyVals() []interface{} {
	return e.keyvals
}
