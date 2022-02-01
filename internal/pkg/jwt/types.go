package jwt

import (
	"context"
	"crypto/rsa"
	"github.com/xmidt-org/ears/pkg/tenant"
	"net/http"
	"sync"
)

const (
	MissingToken           = "missing token"
	InvalidKid             = "invalid jwt kid"
	InvalidSignature       = "invalid jwt signature"
	MissingKid             = "missing jwt kid"
	MissingCapabilities    = "missing jwt capabilities"
	NoMatchingCapabilities = "no matching jwt capabilities"
	MissingClientId        = "missing jwt client id"
	UnauthorizedClientId   = "unauthorized jwt client id"
	UnauthorizedPartnerId  = "unauthorized jwt partner id"
	NoAllowedPartners      = "no allowed partners"
	InvalidSATFormat       = "invalid sat format"
	AllowedResources       = "allowedResources"
	AllowedPartners        = "allowedPartners"
	Capabilities           = "capabilities"
)

type JWTConsumer interface {
	VerifyToken(ctx context.Context, token string, api string, method string, tid *tenant.Id) ([]string, string, error)
}

type (
	DefaultJWTConsumer struct {
		sync.Mutex
		publicKeyEndpoint  string
		keys               map[string]*rsa.PublicKey
		client             *http.Client
		verifier           Verifier
		requireBearerToken bool
		domain             string
		component          string
		adminClientIds     []string
		capabilityPrefixes []string
		tenantStorer       tenant.TenantStorer
	}
	Verifier func(path, method, scope string) bool
)

//401 errors
type UnauthorizedError struct {
	Msg string
}
