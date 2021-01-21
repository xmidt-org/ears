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

package config

import "github.com/xmidt-org/ears/pkg/errs"

func (e *Error) Unwrap() error {
	return e.Err
}

func (e *Error) Error() string {
	return errs.String(
		e,
		nil,
		e.Err,
	)
}

func (e *DataParseError) Unwrap() error {
	return e.Err
}

func (e *DataParseError) Error() string {
	return errs.String(
		e,
		nil,
		e.Err,
	)
}

func (e *InvalidArgumentError) Unwrap() error {
	return e.Err
}

func (e *InvalidArgumentError) Error() string {
	return errs.String(
		e,
		nil,
		e.Err,
	)
}
