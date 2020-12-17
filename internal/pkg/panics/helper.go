package panics

import (
	"bytes"
	"fmt"
	"runtime/debug"
)

const maxStackTraceSize = 8192

type PanicError struct {
	err        string
	stackTrace string
}

func (pe *PanicError) Error() string {
	return pe.err
}

func (pe *PanicError) StackTrace() string {
	return pe.stackTrace
}

func ToError(p interface{}) *PanicError {
	var panicErr string
	switch t := p.(type) {
	case string:
		panicErr = t
	case error:
		panicErr = t.Error()
	default:
		panicErr = fmt.Sprintf("%+v", p)
	}
	stackTrace := bytes.NewBuffer(debug.Stack()).String()
	if len(stackTrace) > maxStackTraceSize {
		stackTrace = stackTrace[:maxStackTraceSize]
	}
	return &PanicError{panicErr, stackTrace}
}
