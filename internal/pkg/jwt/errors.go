package jwt

import "github.com/xmidt-org/ears/pkg/errs"

type JWTAuthError struct {
	Wrapped error
}

func (e *JWTAuthError) Error() string {
	return errs.String("JWTAuthorizationError", nil, e.Wrapped)
}
