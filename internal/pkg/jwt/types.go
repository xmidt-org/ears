package jwt

import (
	"context"
	"crypto/rsa"
	"github.com/xmidt-org/ears/pkg/tenant"
	"net/http"
	"sync"
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
