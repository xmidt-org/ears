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

package app

import "fmt"

// ===================================================================
// Errors
// ===================================================================

// baseError interface makes sure each error
// supports the Go v1.13 error interface
type baseError interface {
	error
	Unwrap() error
}

// Error is a generic error
type Error struct {
	baseError

	Err error
}

func (e *Error) Is(target error) bool {
	return e.Error() == target.Error()
}

func (e *Error) Unwrap() error {
	return e.Err
}

func (e *Error) Error() string {
	if e.Err == nil {
		return "Error: unknown"
	}
	return e.Err.Error()
}

// InvalidOptionError is returned when an invalid option is passed in
type InvalidOptionError struct {
	baseError

	Err error
}

func (e *InvalidOptionError) Is(target error) bool {
	return e.Error() == target.Error()
}

func (e *InvalidOptionError) Unwrap() error {
	return e.Err
}

func (e *InvalidOptionError) Error() string {
	if e.Err == nil {
		return "InvalidOptionError: unknown"
	}

	return fmt.Sprintf("InvalidOptionError: %s", e.Err)
}
