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

package plugin

import (
	"fmt"
	"sort"
	// We're going to be lazy and reference the full Watermill Message object.
)

// Unwrap implement's Go v1.13's error pattern
func (e *Error) Unwrap() error {
	return e.Err
}

// Error implements the standard error interface.  It'll display the code
// as well printing out the Values in key sorted order.
func (e *Error) Error() string {

	msg := fmt.Sprintf("Plugin Error (%v)", e.Code)
	if e.Err != nil {
		msg = msg + fmt.Sprintf(": %s", e.Err.Error())
	}

	keys := []string{}
	for k := range e.Values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		msg = msg + fmt.Sprintf("\n%s: %v", k, e.Values[k])
	}

	return msg
}

// Is compares two Error.  It does a simple string comparison.
// This means that references to memory at different addresses for the same
// key will result in different strings, thus causing `Is` to return false.
func (e *Error) Is(target error) bool {
	return e.Error() == target.Error()
}

// Unwrap implement's Go v1.13's error pattern
func (e *InvalidArgumentError) Unwrap() error {
	return e.Err
}

// Error implements the standard error interface.  It'll display the code
// as well printing out the Values in key sorted order.
func (e *InvalidArgumentError) Error() string {

	msg := fmt.Sprintf("InvalidArgumentError")
	if e.Err != nil {
		msg = msg + fmt.Sprintf(": %s", e.Err.Error())
	}

	return msg
}

// Is compares two Error.  It does a simple string comparison.
// This means that references to memory at different addresses for the same
// key will result in different strings, thus causing `Is` to return false.
func (e *InvalidArgumentError) Is(target error) bool {
	return e.Error() == target.Error()
}

func (e *NotSupportedError) Error() string {
	return "NotSupportedError"
}

func (e *NotSupportedError) Is(target error) bool {
	return e.Error() == target.Error()
}
