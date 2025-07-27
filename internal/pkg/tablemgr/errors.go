// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

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
