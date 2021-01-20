package s3

import (
	"strings"
)

// baseError =============================================================
type baseError struct {
	cause error
}

func (b baseError) Error() string {
	if b.cause != nil {
		// We do this because AWS likes putting in a lot of newlines which we don't want
		return strings.Replace(b.cause.Error(), "\n", " ", -1)
	}
	return "s3 error"
}

func (b baseError) String() string {
	return b.Error()
}

func (b baseError) Cause() error {
	return b.cause
}

// parameterError =========================================================

type parameterError struct {
	baseError
	key   string
	value interface{}
}

func (p parameterError) Key() string {
	return p.key
}

func (p parameterError) Value() interface{} {
	return p.value
}

// requestError =========================================================

type requestError struct {
	baseError
	temporary bool
	url       string
}

func (r requestError) Error() string {
	msg := ""
	if r.cause != nil {
		// We do this because AWS likes putting in a lot of newlines which we don't want
		msg = strings.Replace(r.cause.Error(), "\n", " ", -1)
	}

	if msg != "" {
		msg += ", "
	}

	return msg + "url: " + r.Url()
}

func (r requestError) Temporary() bool {
	return r.temporary
}

func (r requestError) Url() string {
	return r.url
}
