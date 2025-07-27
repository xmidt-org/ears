// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"
	"net/http"

	"github.com/xmidt-org/ears/pkg/errs"
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

type EmptyEventError struct {
	message string
}

func (e *EmptyEventError) Error() string {
	if e.message == "" {
		return "Empty event"
	}
	return e.message
}

func (e *EmptyEventError) StatusCode() int {
	return http.StatusBadRequest
}

type EventTooLargeError struct {
	message string
}

func (e *EventTooLargeError) Error() string {
	if e.message == "" {
		return "Event too large"
	}
	return e.message
}

func (e *EventTooLargeError) StatusCode() int {
	return http.StatusRequestEntityTooLarge
}

type NotFoundError struct {
	message string
}

func (e *NotFoundError) Error() string {
	if e.message == "" {
		return "Item not found"
	}
	return e.message
}

func (e *NotFoundError) StatusCode() int {
	return http.StatusNotFound
}

type NotImplementedError struct {
}

func (e *NotImplementedError) Error() string {
	return "Method not implemented"
}

func (e *NotImplementedError) StatusCode() int {
	return http.StatusNotImplemented
}

type BadRequestError struct {
	message string
	err     error
}

func (e *BadRequestError) Error() string {
	return errs.String("BadRequestError", map[string]interface{}{"message": e.message}, e.err)
}

func (e *BadRequestError) StatusCode() int {
	return http.StatusBadRequest
}

type InternalServerError struct {
	Wrapped error
}

func (e *InternalServerError) Error() string {
	return errs.String("InternalServerError", nil, e.Wrapped)
}

func (e *InternalServerError) StatusCode() int {
	return http.StatusInternalServerError
}
