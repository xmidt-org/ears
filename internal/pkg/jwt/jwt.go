package jwt

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt"
)

func (nfe *UnauthorizedError) Error() string {
	return nfe.Msg
}

func NewJWTConsumer(publicKeyEndpoint string, verifier Verifier, requireBearerToken bool) (JWTConsumer, error) {
	if requireBearerToken {
		if publicKeyEndpoint == "" {
			return nil, errors.New("missing public key endpoint for jwt")
		}
		if verifier == nil {
			return nil, errors.New("missing jwt verifier")
		}
	}
	sc := DefaultJWTConsumer{
		publicKeyEndpoint:  publicKeyEndpoint,
		verifier:           verifier,
		requireBearerToken: requireBearerToken,
		keys:               make(map[string]*rsa.PublicKey),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	return &sc, nil
}

// getKeyData checks cached token and downloads keys from key vault if expired
func (sc *DefaultJWTConsumer) GetKeyData(kid string) (*rsa.PublicKey, error) {
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
	pub, err = jwt.ParseRSAPublicKeyFromPEM(bs)
	if nil != err {
		//ctx.Logger().Error("op", "JWTConsumer.getKeyData", "stage", "decode", "err", err)
		return nil, err
	}
	sc.keys[kid] = pub
	return pub, nil
}

func (sc *DefaultJWTConsumer) VerifyTokenWithClientIds(clientIds []string, domain, component, api, method, token string) ([]string, string, error) {
	if !sc.requireBearerToken {
		return nil, "", nil
	}
	if token == "" {
		return nil, "", &UnauthorizedError{MissingToken}
	}
	partners, sub, capabilities, err := sc.ExtractToken(token)
	if nil != err {
		return nil, "", err
	}
	// verify clientId
	if 0 < len(clientIds) {
		if "" == sub {
			return nil, "", &UnauthorizedError{MissingClientId}
		}
		found := false
		for _, clientId := range clientIds {
			if sub == clientId {
				found = true
				break
			}
		}
		if !found {
			return nil, "", &UnauthorizedError{UnauthorizedClientId}
		}
	}
	// verify capabilities
	for _, cap := range capabilities {
		if sc.IsValid(domain, component, api, method, cap) {
			return partners, sub, nil
		}
	}
	//ctx.Logger().Error("op", "JWTConsumer.VerfiyToken", "stage", "capabilities", "api", api, "method", method, "capabilities", capabilities)
	return nil, "", &UnauthorizedError{NoMatchingCapabilities}
}

func (sc *DefaultJWTConsumer) ExtractToken(token string) ([]string, string, []string, error) {
	var (
		sat *jwt.Token
		err error
	)
	// only allow RS256 to avoid hmac attack, please see details at https://auth0.com/blog/critical-vulnerabilities-in-json-web-token-libraries/
	parser := &jwt.Parser{ValidMethods: []string{"RS256"}}
	sat, err = parser.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if o, has := token.Header["kid"]; has {
			return sc.GetKeyData(o.(string))
		} else {
			return nil, &UnauthorizedError{MissingKid}
		}
	})
	if nil != err {
		if ve, ok := err.(*jwt.ValidationError); ok {
			// key does not exist
			if _, ok := ve.Inner.(*UnauthorizedError); ok {
				return nil, "", nil, ve.Inner
				// token expired or not valid yet
			} else {
				return nil, "", nil, &UnauthorizedError{ve.Error()}
			}
		}
		// internal error
		//ctx.Logger().Error("op", "JWTConsumer.VerfiyToken", "stage", "validate", "err", err)
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
	// TODO: return error if partners cannot be found
	var partners []string
	if res, ok := claims["allowedResources"].(map[string]interface{}); ok {
		if os, ok := res["allowedPartners"].([]interface{}); ok {
			for _, o := range os {
				if p, ok := o.(string); ok {
					partners = append(partners, p)
				}
			}
		}
	}
	os, ok := claims["capabilities"].([]interface{})
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
func (sc *DefaultJWTConsumer) IsValid(domain, component, api, method, cap string) bool {
	//SAT requires prefix which is not related to the capabilities, remove it
	cap = strings.TrimPrefix(cap, SATPrefix)
	cap = strings.TrimPrefix(cap, SATPrefix1)
	ss := strings.Split(cap, ":")
	if len(ss) < 4 {
		return false
	}
	idx := 0
	if len(ss) > 4 {
		idx = 1
	}
	if domain != ss[idx] {
		return false
	}
	if component != ss[idx+1] {
		return false
	}
	return sc.verifier(api, method, ss[idx+2]+":"+ss[idx+3])
}
