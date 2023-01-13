package jwt

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/xmidt-org/ears/pkg/tenant"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt"
)

func (nfe *UnauthorizedError) Error() string {
	return nfe.Msg
}

func NewJWTConsumer(publicKeyEndpoint string, verifier Verifier, requireBearerToken bool, domain string, component string, adminClientIds []string, capabilityPrefixes []string, tenantStorer tenant.TenantStorer) (JWTConsumer, error) {
	if requireBearerToken {
		if publicKeyEndpoint == "" {
			return nil, errors.New("missing public key endpoint for jwt")
		}
		if verifier == nil {
			return nil, errors.New("missing jwt verifier")
		}
		if domain == "" || component == "" {
			return nil, errors.New("missing jwt domain or component")
		}
	}
	sc := DefaultJWTConsumer{
		publicKeyEndpoint:  publicKeyEndpoint,
		verifier:           verifier,
		requireBearerToken: requireBearerToken,
		keys:               make(map[string]*rsa.PublicKey),
		domain:             domain,
		component:          component,
		adminClientIds:     adminClientIds,
		capabilityPrefixes: capabilityPrefixes,
		tenantStorer:       tenantStorer,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	return &sc, nil
}

// getKeyData checks cached token and downloads keys from key vault if expired
func (sc *DefaultJWTConsumer) getPublicKey(kid string) (*rsa.PublicKey, error) {
	sc.Lock()
	defer sc.Unlock()
	pub, has := sc.keys[kid]
	if has {
		return pub, nil
	}
	// fetch from key vault
	req, _ := http.NewRequest("GET", sc.publicKeyEndpoint+"/"+kid, nil)
	resp, err := sc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bs, err := io.ReadAll(resp.Body)
	if nil != err {
		if resp.StatusCode == 404 {
			return nil, &UnauthorizedError{InvalidKid}
		} else {
			return nil, err
		}
	}
	if 200 != resp.StatusCode {
		//ctx.Logger().Error("op", "JWTConsumer.getKeyData", "stage", "client.DoRequest", "status", status)
		return nil, fmt.Errorf("Invalid status code: %d\n", resp.StatusCode)
	}
	if strings.Contains(sc.publicKeyEndpoint, "v2") {
		pub, err = convertJwkToPublicKey(bs)
		if err != nil {
			return nil, err
		}
	} else {
		pub, err = jwt.ParseRSAPublicKeyFromPEM(bs)
		if nil != err {
			return nil, err
		}
	}
	sc.keys[kid] = pub
	return pub, nil
}

func convertJwkToPublicKey(input []byte) (*rsa.PublicKey, error) {
	jwk := map[string]string{}
	json.Unmarshal([]byte(input), &jwk)

	// decode the base64 bytes for n
	nb, err := base64.RawURLEncoding.DecodeString(jwk["n"])
	if err != nil {
		return nil, err
	}

	e := 0
	// The default exponent is usually 65537, so just compare the
	// base64 for [1,0,1] or [0,1,0,1]
	if jwk["e"] == "AQAB" || jwk["e"] == "AAEAAQ" {
		e = 65537
	} else {
		// need to decode "e" as a big-endian int
		return nil, errors.New("need to deocde e:" + jwk["e"])
	}

	pk := &rsa.PublicKey{
		N: new(big.Int).SetBytes(nb),
		E: e,
	}
	return pk, nil
}


// VerifyToken validates a give token for an optional list of clientIds / "subjects" (use empty array if any client id ok) and
// for a given domain, component, api and method (api and method are encoded in the claim)
func (sc *DefaultJWTConsumer) VerifyToken(ctx context.Context, token string, api string, method string, tid *tenant.Id) ([]string, string, error) {
	if !sc.requireBearerToken {
		return nil, "", nil
	}
	if token == "" {
		return nil, "", &UnauthorizedError{MissingToken}
	}
	partners, sub, capabilities, err := sc.extractToken(token)
	if nil != err {
		return nil, "", err
	}
	// verify clientId and partner
	if "" == sub {
		return nil, "", &UnauthorizedError{MissingClientId}
	}
	// check if this is an admin client
	if 0 < len(sc.adminClientIds) {
		for _, clientId := range sc.adminClientIds {
			if sub == clientId {
				return partners, sub, nil
			}
		}
	}
	// otherwise check if this is an app client
	foundClientId := false
	tenantConfig, err := sc.tenantStorer.GetConfig(ctx, *tid)
	if err != nil {
		return nil, "", err
	}
	for _, cid := range tenantConfig.ClientIds {
		if sub == cid {
			foundClientId = true
			break
		}
	}
	if !foundClientId {
		return nil, "", &UnauthorizedError{UnauthorizedClientId}
	}
	// verify allowed partners if any are given in token
	if len(partners) > 0 {
		foundAllowedPartner := false
		for _, partner := range partners {
			if partner == tid.OrgId {
				foundAllowedPartner = true
				break
			}
		}
		if !foundAllowedPartner {
			return nil, "", &UnauthorizedError{UnauthorizedPartnerId}
		}
	}
	// verify capabilities
	for _, cap := range capabilities {
		if sc.isValid(api, method, cap) {
			return partners, sub, nil
		}
	}
	return nil, "", &UnauthorizedError{NoMatchingCapabilities}
}

// extractToken returns array of partner strings (from claims.allowedResources.allowedPartners),
// subject string (aka clientId), an array of capabilities and finally an error which is nil if the
// token is valid
func (sc *DefaultJWTConsumer) extractToken(token string) ([]string, string, []string, error) {
	var (
		sat *jwt.Token
		err error
	)
	// only allow RS256 to avoid hmac attack, please see details at https://auth0.com/blog/critical-vulnerabilities-in-json-web-token-libraries/
	parser := &jwt.Parser{ValidMethods: []string{"RS256", "ES256"}}
	sat, err = parser.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if o, has := token.Header["kid"]; has {
			return sc.getPublicKey(o.(string))
		} else {
			return nil, &UnauthorizedError{MissingKid}
		}
	})
	if nil != err {
		var ve *jwt.ValidationError
		if errors.As(err, &ve) {
			var veInner *UnauthorizedError
			if errors.As(ve.Inner, &veInner) {
				return nil, "", nil, veInner
				// token expired or not valid yet
			} else {
				return nil, "", nil, &UnauthorizedError{ve.Error()}
			}
		}
		// internal error
		return nil, "", nil, err
	}
	if !sat.Valid {
		return nil, "", nil, &UnauthorizedError{InvalidSignature}
	}
	claims, ok := sat.Claims.(jwt.MapClaims)
	if !ok {
		return nil, "", nil, &UnauthorizedError{InvalidSATFormat}
	}
	sub, _ := claims["sub"].(string)
	// get partners
	var partners []string
	if res, ok := claims[AllowedResources].(map[string]interface{}); ok {
		if os, ok := res[AllowedPartners].([]interface{}); ok {
			for _, o := range os {
				if p, ok := o.(string); ok {
					partners = append(partners, p)
				}
			}
		}
	}
	if len(partners) == 0 {
		return nil, "", nil, &UnauthorizedError{NoAllowedPartners}
	}
	os, ok := claims[Capabilities].([]interface{})
	if !ok {
		return nil, "", nil, &UnauthorizedError{MissingCapabilities}
	}
	caps := make([]string, 0, len(os))
	for _, o := range os {
		if s, ok := o.(string); ok {
			caps = append(caps, s)
		}
	}
	return partners, sub, caps, nil
}

// isValid verifies capability in two formats:
//  * <domain>:<component>:<api>:<method>
//  * <ignore>:<domain>:<component>:<api>:<method>
func (sc *DefaultJWTConsumer) isValid(api, method, cap string) bool {
	// SAT requires prefix which is not related to the capabilities, remove it
	for _, prefix := range sc.capabilityPrefixes {
		cap = strings.TrimPrefix(cap, prefix)
	}
	ss := strings.Split(cap, ":")
	if len(ss) < 4 {
		return false
	}
	idx := 0
	if len(ss) > 4 {
		idx = 1
	}
	if sc.domain != ss[idx] {
		return false
	}
	if sc.component != ss[idx+1] {
		return false
	}
	return sc.verifier(api, method, ss[idx+2]+":"+ss[idx+3])
}
