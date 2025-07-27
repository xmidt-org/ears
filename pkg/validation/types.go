// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package validation

type Validator interface {
	Validate() error
}

type SchemaProvider interface {
	Schema() string
}

type Error struct {
	// TODO
	Err error
}

type Errors struct {
	Errs []error
}

type ProcessingError struct {
	Err error
}
