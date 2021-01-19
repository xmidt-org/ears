package validation

import (
	"strings"

	"github.com/xmidt-org/ears/pkg/errs"
)

func (e *Errors) Unwrap() error {
	if len(e.Errs) > 0 {
		return e.Errs[0]
	}
	return nil
}

func (e *Errors) Error() string {
	var msgs = []string{}
	for _, err := range e.Errs {
		msgs = append(msgs, err.Error())
	}

	return strings.Join(msgs, ", ")
}

func (e *Error) Unwrap() error {
	return e.Err
}

func (e *Error) Error() string {
	return errs.String(
		nil,
		nil,
		e.Err,
	)
}

func (e *ProcessingError) Unwrap() error {
	return e.Err
}

func (e *ProcessingError) Error() string {
	return errs.String(
		e,
		nil,
		e.Err,
	)
}
