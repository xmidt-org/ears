package jwt

import "github.com/xmidt-org/ears/pkg/errs"

const (
	MissingToken           = "missing token"
	InvalidKid             = "invalid jwt kid"
	InvalidSignature       = "invalid jwt signature"
	MissingKid             = "missing jwt kid"
	MissingCapabilities    = "missing jwt capabilities"
	NoMatchingCapabilities = "no matching jwt capabilities"
	MissingClientId        = "missing jwt client id"
	MissingTenantId        = "missing tenant id"
	UnauthorizedClientId   = "unauthorized jwt client id"
	UnauthorizedPartnerId  = "unauthorized jwt partner id"
	NoAllowedPartners      = "no allowed partners"
	InvalidSATFormat       = "invalid sat format"
	AllowedResources       = "allowedResources"
	AllowedPartners        = "allowedPartners"
	Capabilities           = "capabilities"
)

type JWTAuthError struct {
	Wrapped error
}

func (e *JWTAuthError) Error() string {
	return errs.String("JWTAuthorizationError", nil, e.Wrapped)
}
