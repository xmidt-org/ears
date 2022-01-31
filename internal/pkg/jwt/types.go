package jwt

import (
	"crypto/rsa"
	"github.com/xmidt-org/ears/pkg/tenant"
	"net/http"
	"sync"
)

const (
	SATPrefix  = "x1:capabilities:"
	SATPrefix1 = "x1:"

	MissingToken           = "missing token"
	InvalidKid             = "invalid jwt kid"
	InvalidSignature       = "invalid jwt signature"
	MissingKid             = "missing jwt kid"
	MissingCapabilities    = "missing jwt capabilities"
	NoMatchingCapabilities = "no matching jwt capabilities"
	MissingClientId        = "missing jwt client id"
	UnauthorizedClientId   = "unauthorized jwt client id or disallowed partner"
	NoAllowedPartners      = "no allowed partners"
	TokenExpired           = "jwt token is expired"
	TokenNotValidYet       = "jwt token is not valid yet"
	InvalidSATFormat       = "invalid sat format"
)

type JWTConsumer interface {
	VerifyToken(token string, api string, method string, tid *tenant.Id) ([]string, string, error)
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
	}
	Verifier func(path, method, scope string) bool
)

//401 errors
type UnauthorizedError struct {
	Msg string
}
