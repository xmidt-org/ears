// Copyright 2021 Comcast Cable Communications Management, LLC
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

package s3

import (
	"strings"
)

// baseError =============================================================
type baseError struct {
	cause error
}

func (b baseError) Error() string {
	if b.cause != nil {
		// We do this because AWS likes putting in a lot of newlines which we don't want
		return strings.Replace(b.cause.Error(), "\n", " ", -1)
	}
	return "s3 error"
}

func (b baseError) String() string {
	return b.Error()
}

func (b baseError) Cause() error {
	return b.cause
}

// parameterError =========================================================

type parameterError struct {
	baseError
	key   string
	value interface{}
}

func (p parameterError) Key() string {
	return p.key
}

func (p parameterError) Value() interface{} {
	return p.value
}

// requestError =========================================================

type requestError struct {
	baseError
	temporary bool
	url       string
}

func (r requestError) Error() string {
	msg := ""
	if r.cause != nil {
		// We do this because AWS likes putting in a lot of newlines which we don't want
		msg = strings.Replace(r.cause.Error(), "\n", " ", -1)
	}

	if msg != "" {
		msg += ", "
	}

	return msg + "url: " + r.Url()
}

func (r requestError) Temporary() bool {
	return r.temporary
}

func (r requestError) Url() string {
	return r.url
}
