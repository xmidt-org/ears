// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package plugin

import "github.com/xmidt-org/ears/pkg/errs"

func (e *OptionError) Unwrap() error {
	return e.Err
}

func (e *OptionError) Error() string {
	return errs.String(
		"OptionError",
		map[string]interface{}{
			"message": e.Message,
		},
		e.Err,
	)
}

func (e *RegistrationError) Unwrap() error {
	return e.Err
}

func (e *RegistrationError) Error() string {
	return errs.String(
		"RegistrationError",
		map[string]interface{}{
			"message": e.Message,
			"plugin":  e.Name,
			"name":    e.Plugin,
		},
		e.Err,
	)
}

func (e *UnregistrationError) Unwrap() error {
	return e.Err
}

func (e *UnregistrationError) Error() string {
	return errs.String(
		"UnregistrationError",
		map[string]interface{}{
			"message": e.Message,
		},
		e.Err,
	)
}

func (e *NotRegisteredError) Unwrap() error {
	return nil
}

func (e *NotRegisteredError) Error() string {
	return errs.String("NotRegisteredError", nil, nil)
}
