package panic

import (
	"bytes"
	"errors"
	"fmt"
	"runtime/debug"
)

const maxStackTraceSize = 8192

func GetPanicInfo(p interface{}) (error, string) {
	var panicErr error
	switch t := p.(type) {
	case string:
		panicErr = errors.New(t)
	case error:
		panicErr = t
	default:
		panicErr = errors.New(fmt.Sprintf("%+v", p))
	}
	stackTrace := bytes.NewBuffer(debug.Stack()).String()
	if len(stackTrace) > maxStackTraceSize {
		stackTrace = stackTrace[:maxStackTraceSize]
	}
	return panicErr, stackTrace
}
