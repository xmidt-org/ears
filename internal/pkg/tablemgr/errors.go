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

package tablemgr

import "github.com/xmidt-org/ears/pkg/errs"

// standard set of errors for route manager

type RouteValidationError struct {
	Wrapped error
}

func (e *RouteValidationError) Error() string {
	return errs.String("RouteRegistrationError", nil, e.Wrapped)
}

type RouteRegistrationError struct {
	Wrapped error
}

func (e *RouteRegistrationError) Error() string {
	return errs.String("RouteRegistrationError", nil, e.Wrapped)
}

type RouteNotFoundError struct {
	Id string
}

func (e *RouteNotFoundError) Error() string {
	return errs.String("RouteNotFoundError", map[string]interface{}{"id": e.Id}, nil)
}

type BadConfigError struct {
	Wrapped error
}

func (e *BadConfigError) Error() string {
	return errs.String("BadConfigError", nil, e.Wrapped)
}

type InternalStorageError struct {
	Wrapped error
}

func (e *InternalStorageError) Error() string {
	return errs.String("InternalStorageError", nil, e.Wrapped)
}
