package ratelimit

import "github.com/xmidt-org/ears/pkg/errs"

type InvalidUnitError struct {
	BadUnit int
}

func (e *InvalidUnitError) Error() string {
	return errs.String("InvalidUnitError", map[string]interface{}{"unit": e.BadUnit}, nil)
}

type LimitReached struct {
	Limit int
}

func (e *LimitReached) Error() string {
	return errs.String("LimitReached", map[string]interface{}{"limit": e.Limit}, nil)
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
