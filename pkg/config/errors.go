// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

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
