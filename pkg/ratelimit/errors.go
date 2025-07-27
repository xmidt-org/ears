// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package ratelimit

import "github.com/xmidt-org/ears/pkg/errs"

type InvalidUnitError struct {
	BadUnit int
}

func (e *InvalidUnitError) Error() string {
	return errs.String("InvalidUnitError", map[string]interface{}{"unit": e.BadUnit}, nil)
}

type LimitReached struct {
}

func (e *LimitReached) Error() string {
	return errs.String("LimitReached", nil, nil)
}

type BackendError struct {
	Source error
}

func (e *BackendError) Error() string {
	return errs.String("BackendError", nil, e.Source)
}

func (e *BackendError) Unwrap() error {
	return e.Source
}

type ContextCancelled struct {
}

func (e *ContextCancelled) Error() string {
	return errs.String("ContextCancelled", nil, nil)
}
