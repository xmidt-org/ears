// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package panics

const maxStackTraceSize = 8192

type PanicError struct {
	err        string
	stackTrace string
}

func (pe *PanicError) Error() string {
	return pe.err
}

func (pe *PanicError) StackTrace() string {
	return pe.stackTrace
}
