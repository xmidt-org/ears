// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"github.com/xmidt-org/ears/pkg/errs"
)

type EmptyCmdArgumentError struct {
}

func (e *EmptyCmdArgumentError) Error() string {
	return errs.String("EmptyCmdArgumentError", nil, nil)
}

type NoDefaultSetError struct {
}

func (e *NoDefaultSetError) Error() string {
	return errs.String("NoDefaultSetError", nil, nil)
}

type ConfigNotSupportedProtocolError struct {
	protocol string
}

func (e *ConfigNotSupportedProtocolError) Error() string {
	return errs.String("ConfigNotSupportedProtocolError", map[string]interface{}{"protocol": e.protocol}, nil)
}

type ConfigError struct {
	Err  error
	path string
}

func (c *ConfigError) Error() string {
	return errs.String("ConfigError", map[string]interface{}{"path": c.path}, c.Err)
}

func (c *ConfigError) Unwrap() error {
	return c.Err
}
