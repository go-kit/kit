package log

import (
	"fmt"
	"path/filepath"
	"runtime"
)

type Valuer interface {
	Value() interface{}
}

var Caller = ValuerFunc(func() interface{} {
	_, file, line, _ := runtime.Caller(5)
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
})

type ValuerFunc func() interface{}

func (f ValuerFunc) Value() interface{} { return f() }
