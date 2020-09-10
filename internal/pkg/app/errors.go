package app

import "fmt"

// ===================================================================
// Errors
// ===================================================================

// baseError interface makes sure each error
// supports the Go v1.13 error interface
type baseError interface {
	error
	Unwrap() error
}

// Error is a generic error
type Error struct {
	baseError

	Err error
}

// InvalidOptionError is returned when an invalid option is passed in
type InvalidOptionError struct {
	baseError

	Err error
}

// NotInitializedError is returned when neither command or script have been set
type NotInitializedError struct {
	baseError
}

func (e *Error) Unwrap() error {
	return e.Err
}

func (e *Error) Error() string {
	if e.Err == nil {
		return "Error: unknown"
	}
	return e.Err.Error()
}

func (e *Error) Is(target error) bool {
	return e.Error() == target.Error()
}

func (e *InvalidOptionError) Unwrap() error {
	return e.Err
}

func (e *InvalidOptionError) Error() string {
	if e.Err == nil {
		return "InvalidOptionError: unknown"
	}

	return fmt.Sprintf("InvalidOptionError: %s", e.Err)
}

func (e *NotInitializedError) Unwrap() error {
	return nil
}

func (e *NotInitializedError) Error() string {
	return "NotInitializedError"
}
