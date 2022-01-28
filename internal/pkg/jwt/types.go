package jwt

import "crypto/rsa"

type JWTConsumer interface {
	GetKeyData(kid string) (*rsa.PublicKey, error)
	VerifyTokenWithClientIds(clientIds []string, domain, component, api, method, token string) ([]string, string, error)
	ExtractToken(token string) ([]string, string, []string, error)
	IsValid(domain, component, api, method, cap string) bool
}
