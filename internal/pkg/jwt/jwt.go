package jwt

import (
	"context"
	"crypto/rsa"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	jwt "github.com/golang-jwt/jwt"
)

const (
	SATPrefix  = "x1:capabilities:"
	SATPrefix1 = "x1:"

	InvalidKid           = "invalid sat kid"
	InvalidSignature     = "invalid sat signature"
	MissingKid           = "missing sat kid"
	MissingCapabilities  = "missing sat capabilities"
	NoMatchCapabilities  = "no matching sat capabilities"
	MissingClientId      = "missing sat client id"
	UnauthorizedClientId = "unauthorized sat client id"
	TokenExpired         = "sat token is expired"
	TokenNotValidYet     = "sat token is not valid yet"
	InvalidSATFormat     = "invalid sat format"
)

type (
	JWTConsumer struct {
		sync.Mutex
		keyVault string
		keys     map[string]*rsa.PublicKey
		client   *http.Client
		verifier Verifier
	}
	Verifier func(path, method, scope string) bool
)

//401 errors
type UnauthorizedError struct {
	Msg string
}

func (nfe *UnauthorizedError) Error() string {
	return nfe.Msg
}

// InitConsumer inits JWTConsumer
func InitConsumer(ctx context.Context, keyVault string, verifier Verifier) (*JWTConsumer, error) {
	sc := JWTConsumer{
		keyVault: keyVault,
		verifier: verifier,
		keys:     make(map[string]*rsa.PublicKey),
		client: &http.Client{
			Timeout: time.Duration(30 * time.Second),
		},
	}
	return &sc, nil
}

// getKeyData checks cached token and downloads keys from key vault if expired
// TODO: implement key revocation flow
func (sc *JWTConsumer) getKeyData(kid string) (*rsa.PublicKey, error) {
	sc.Lock()
	defer sc.Unlock()
	pub, has := sc.keys[kid]
	if has {
		return pub, nil
	}
	// fetch from key vault
	//status, bs, err := sc.client.DoRequest(ctx, "GET", sc.keyVault+"/"+kid, nil, nil)
	req, err := http.NewRequest("GET", sc.keyVault+"/"+kid, nil)
	//req.Header.Add("", "")
	resp, err := sc.client.Do(req)
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
	bs, err := io.ReadAll(resp.Body)
	if nil != err {
		//ctx.Logger().Error("op", "JWTConsumer.getKeyData", "stage", "client.DoRequest", "status", status, "err", err)
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

func (sc *JWTConsumer) verifyTokenWithClientIds(clientIds []string, domain, component, api, method, token string) ([]string, string, error) {
	partners, sub, capabilities, err := sc.extractToken(token)
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
		if sc.isValid(domain, component, api, method, cap) {
			return partners, sub, nil
		}
	}
	//ctx.Logger().Error("op", "JWTConsumer.VerfiyToken", "stage", "capabilities", "api", api, "method", method, "capabilities", capabilities)
	return nil, "", &UnauthorizedError{NoMatchCapabilities}
}

func (sc *JWTConsumer) extractToken(token string) ([]string, string, []string, error) {
	var (
		sat *jwt.Token
		err error
	)
	//only allow RS256 to avoid hmac attack, please see details at https://auth0.com/blog/2015/03/31/critical-vulnerabilities-in-json-web-token-libraries/
	parser := &jwt.Parser{ValidMethods: []string{"RS256"}}
	sat, err = parser.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if o, has := token.Header["kid"]; has {
			return sc.getKeyData(o.(string))
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
func (sc *JWTConsumer) isValid(domain, component, api, method, cap string) bool {
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
