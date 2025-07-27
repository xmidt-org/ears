// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package filter

import "github.com/xmidt-org/ears/pkg/errs"

func (e *InvalidConfigError) Unwrap() error {
	return e.Err
}

func (e *InvalidConfigError) Error() string {
	return errs.String("InvalidConfigError", nil, e.Err)
}

func (e *InvalidArgumentError) Unwrap() error {
	return e.Err
}

func (e *InvalidArgumentError) Error() string {
	return errs.String("InvalidArgumentError", nil, e.Err)
}
