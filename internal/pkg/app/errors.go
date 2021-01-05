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

import (
	"fmt"
	"net/http"
)

// InvalidOptionError is returned when an invalid option is passed in
type InvalidOptionError struct {
	Option string
}

func (e *InvalidOptionError) Error() string {
	return fmt.Sprintf("InvalidOptionError (Option=%s)", e.Option)
}

// Rest API errors
type ApiError interface {
	error
	StatusCode() int
}

type NotFoundError struct {
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("Item not found")
}

func (e *NotFoundError) StatusCode() int {
	return http.StatusNotFound
}

type NotImplementedError struct {
}

func (e *NotImplementedError) Error() string {
	return fmt.Sprintf("Method not implemented")
}

func (e *NotImplementedError) StatusCode() int {
	return http.StatusNotImplemented
}
