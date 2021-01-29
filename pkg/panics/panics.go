// Copyright 2020 Comcast Cable Communications Management, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
