// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package panics

import (
	"bytes"
	"fmt"
	"runtime/debug"
)

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
