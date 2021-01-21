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
	"github.com/xmidt-org/ears/pkg/errs"
)

// Unwrap implement's Go v1.13's error pattern
func (e *Error) Unwrap() error {
	return e.Err
}

// Error implements the standard error interface.
//
// Output:
//   plugin error (code=42): wrapped Err.Error() goes here
//
func (e *Error) Error() string {
	return errs.String(
		"plugin error",
		map[string]interface{}{
			"code": e.Code,
		},
		e.Err,
	)
}

// Is compares two Error.  It does a simple string comparison.
// This means that references to memory at different addresses for the same
// key will result in different strings, thus causing `Is` to return false.
func (e *Error) Is(target error) bool {
	return e.Error() == target.Error()
}

// Unwrap implement's Go v1.13's error pattern
func (e *InvalidConfigError) Unwrap() error {
	return e.Err
}

// Error implements the standard error interface.  It'll display the code
// as well printing out the Values in key sorted order.
func (e *InvalidConfigError) Error() string {
	return errs.String("InvalidConfigError", nil, e.Err)
}

// Is compares two Error.  It does a simple string comparison.
// This means that references to memory at different addresses for the same
// key will result in different strings, thus causing `Is` to return false.
func (e *InvalidConfigError) Is(target error) bool {
	return e.Error() == target.Error()
}

// Error implements the standard error interface.  It'll display the code
// as well printing out the Values in key sorted order.
func (e *NotSupportedError) Error() string {
	return errs.String("NotSupportedError", nil, nil)
}

// Is compares two Error.  It does a simple string comparison.
// This means that references to memory at different addresses for the same
// key will result in different strings, thus causing `Is` to return false.
func (e *NotSupportedError) Is(target error) bool {
	return e.Error() == target.Error()
}

// TODO
func (e *OptionError) Error() string {
	return errs.String("OptionError", nil, e.Err)
}

func (e *NilPluginError) Error() string {
	return errs.String("NilPluginError", nil, nil)
}
