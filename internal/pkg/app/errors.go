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

func (e *Error) Is(target error) bool {
	return e.Error() == target.Error()
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

// InvalidOptionError is returned when an invalid option is passed in
type InvalidOptionError struct {
	baseError

	Err error
}

func (e *InvalidOptionError) Is(target error) bool {
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
