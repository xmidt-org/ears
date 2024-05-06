package appsecret

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
)

// ConfigVault provides secrets from ears app configuration
type ConfigVault struct {
	config config.Config
}

type (
	SatToken struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		Scope       string `json:"scope"`
		TokenType   string `json:"token_type"`
		ExpiresAt   int64  `json:"timestamp"`
	}

	Autobahn struct {
		Basic
		AppRole    string `json:"appRole,omitempty"`
		PrivateKey string `json:"privateKey,omitempty"`
		PublicKey  string `json:"publicKey,omitempty"`
		SSHRole    string `json:"sshRole,omitempty"`
		Token      string `json:"token,omitempty"`
	}

	Basic struct {
		Id     string `json:"id"`
		Secret string `json:"secret"`
	}

	Oauth struct {
		Basic
		Issuer         string `json:"issuer"`
		GrantType      string `json:"grantType,omitempty"`
		Scope          string `json:"scope,omitempty"`
		Level          string `json:"level,omitempty"`
		SecretLocation int    `json:"secretLocation,omitempty"`
		TokenType      string `json:"tokenType,omitempty"`
	}

	Data struct {
		Autobahn *Autobahn `json:"autobahn,omitempty"`
		Basic    *Basic    `json:"basic,omitempty"`
		Generic  *Basic    `json:"generic,omitempty"`
		SSH      *Basic    `json:"ssh,omitempty"`
		OAuth1   *Oauth    `json:"oauth1,omitempty"`
		OAuth2   *Oauth    `json:"oauth2,omitempty"`
		SAT      *Oauth    `json:"sat,omitempty"`
	}

	Credential struct {
		Name string `json:"name"`
		Type string `json:"type"`
		Data Data   `json:"data,omitempty"`
		Ttl  int32  `json:"ttl,omitempty"`
	}

	CredentialResponse struct {
		Item Credential `json:"item"`
	}
)

var satToken SatToken

func NewConfigVault(config config.Config) secret.Vault {
	return &ConfigVault{config: config}
}

func (v *ConfigVault) Secret(ctx context.Context, key string) string {
	if strings.HasPrefix(key, secret.ProtocolSecret) {
		configKey := "ears.secrets." + key[len(secret.ProtocolSecret):]
		return v.config.GetString(configKey)
	} else if strings.HasPrefix(key, secret.ProtocolCredential) {
		// credentials API must always be tenant specific
		return key
	} else {
		return key
	}
}

func (v *ConfigVault) GetConfig(ctx context.Context, key string) string {
	return v.config.GetString(key)
}

type TenantConfigVault struct {
	parentVault  secret.Vault
	tid          tenant.Id
	tenantStorer tenant.TenantStorer
	httpClient   *http.Client
	logger       *zerolog.Logger
}

func NewTenantConfigVault(tid tenant.Id, parentVault secret.Vault, tenantStorer tenant.TenantStorer, logger *zerolog.Logger) secret.Vault {
	tcv := &TenantConfigVault{
		parentVault:  parentVault,
		tid:          tid,
		tenantStorer: tenantStorer,
		logger:       logger,
	}
	return tcv
}

func (v *TenantConfigVault) getSatBearerToken(ctx context.Context) (string, error) {
	if time.Now().Unix() >= satToken.ExpiresAt {
		satToken = SatToken{}
	}
	if satToken.AccessToken != "" {
		return satToken.TokenType + " " + satToken.AccessToken, nil
	}
	if v.httpClient == nil {
		v.httpClient = &http.Client{
			Timeout: 10 * time.Second,
		}
	}
	SATUrl := v.Secret(ctx, "secret://SATUrl")
	if SATUrl == "" {
		return "", errors.New("missing SAT URL")
	}
	req, err := http.NewRequest("POST", SATUrl, nil)
	if err != nil {
		return "", err
	}
	tenantConfig, err := v.tenantStorer.GetConfig(ctx, v.tid)
	if err != nil {
		return "", err
	}
	clientId := ""
	if tenantConfig.ClientId != "" {
		clientId = tenantConfig.ClientId
	} else if len(tenantConfig.ClientIds) > 0 {
		clientId = tenantConfig.ClientIds[0]
	}
	if clientId == "" {
		return "", errors.New("tenant has no client IDs")
	}
	if tenantConfig.ClientSecret == "" {
		return "", errors.New("tenant has no client secret")
	}
	req.Header.Add("X-Client-Id", clientId)
	req.Header.Add("X-Client-Secret", tenantConfig.ClientSecret)
	req.Header.Add("Cache-Control", "no-cache")
	resp, err := v.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	err = json.Unmarshal(buf, &satToken)
	if err != nil {
		return "", err
	}
	satToken.ExpiresAt = time.Now().Unix() + int64(satToken.ExpiresIn)
	return satToken.TokenType + " " + satToken.AccessToken, nil
}

func (v *TenantConfigVault) GetConfig(ctx context.Context, key string) string {
	return v.parentVault.GetConfig(ctx, key)
}

func (v *TenantConfigVault) getCredential(ctx context.Context, key string, credentialType string, field string) (*Credential, error) {
	token, err := v.getSatBearerToken(ctx)
	if err != nil {
		return nil, err
	}
	if v.httpClient == nil {
		v.httpClient = &http.Client{
			Timeout: 10 * time.Second,
		}
	}
	env := v.GetConfig(ctx, "ears.env") + "."
	if env == "prod." {
		env = ""
	}
	url := v.Secret(ctx, "CredentialUrl")
	if url == "" {
		return nil, errors.New("missing credential URL")
	}
	url = strings.Replace(url, "{{app}}", v.tid.AppId, -1)
	url = strings.Replace(url, "{{key}}", key, -1)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", token)
	req.Header.Add("Partner-Id", v.tid.OrgId)
	req.Header.Add("Cache-Control", "no-cache")
	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var cred CredentialResponse
	err = json.Unmarshal(buf, &cred)
	if err != nil {
		return nil, err
	}
	return &cred.Item, nil
}

func (c *Credential) GetSecret() (string, string) {
	if c.Data.Generic.Secret != "" {
		return "generic", c.Data.Generic.Secret
	} else if c.Data.SAT.Secret != "" {
		return "sat", c.Data.SAT.Secret
	} else if c.Data.SSH.Secret != "" {
		return "ssh", c.Data.SSH.Secret
	} else if c.Data.Basic.Secret != "" {
		return "basic", c.Data.Basic.Secret
	} else if c.Data.OAuth1.Secret != "" {
		return "oauth1", c.Data.OAuth1.Secret
	} else if c.Data.OAuth2.Secret != "" {
		return "oauth2", c.Data.OAuth2.Secret
	} else if c.Data.Autobahn.Secret != "" {
		return "autobahn", c.Data.Autobahn.Secret
	}
	return "", ""
}

func (c *Credential) GetId() (string, string) {
	if c.Data.Generic.Id != "" {
		return "generic", c.Data.Generic.Id
	} else if c.Data.SAT.Id != "" {
		return "sat", c.Data.SAT.Id
	} else if c.Data.SSH.Id != "" {
		return "ssh", c.Data.SSH.Id
	} else if c.Data.Basic.Id != "" {
		return "basic", c.Data.Basic.Id
	} else if c.Data.OAuth1.Id != "" {
		return "oauth1", c.Data.OAuth1.Id
	} else if c.Data.OAuth2.Id != "" {
		return "oauth2", c.Data.OAuth2.Id
	} else if c.Data.Autobahn.Id != "" {
		return "autobahn", c.Data.Autobahn.Id
	}
	return "", ""
}

func (c *Credential) GetGrantType() (string, string) {
	if c.Data.SAT.GrantType != "" {
		return "sat", c.Data.SAT.GrantType
	} else if c.Data.OAuth1.GrantType != "" {
		return "oauth1", c.Data.OAuth1.GrantType
	} else if c.Data.OAuth2.GrantType != "" {
		return "oauth2", c.Data.OAuth2.GrantType
	}
	return "", ""
}

func (c *Credential) GetScope() (string, string) {
	if c.Data.SAT.Scope != "" {
		return "sat", c.Data.SAT.Scope
	} else if c.Data.OAuth1.Scope != "" {
		return "oauth1", c.Data.OAuth1.Scope
	} else if c.Data.OAuth2.Scope != "" {
		return "oauth2", c.Data.OAuth2.Scope
	}
	return "", ""
}

func (c *Credential) GetIssuer() (string, string) {
	if c.Data.SAT.Issuer != "" {
		return "sat", c.Data.SAT.Issuer
	} else if c.Data.OAuth1.Issuer != "" {
		return "oauth1", c.Data.OAuth1.Issuer
	} else if c.Data.OAuth2.Issuer != "" {
		return "oauth2", c.Data.OAuth2.Issuer
	}
	return "", ""
}

func (v *TenantConfigVault) Secret(ctx context.Context, key string) string {
	if strings.HasPrefix(key, secret.ProtocolSecret) {
		configKey := secret.ProtocolSecret + v.tid.OrgId + "." + v.tid.AppId + "." + key[len(secret.ProtocolSecret):]
		val := v.parentVault.Secret(ctx, configKey)
		if val != "" {
			return val
		}
		// try again with global key/secrets
		configKey = secret.ProtocolSecret + "all.all." + key[len(secret.ProtocolSecret):]
		return v.parentVault.Secret(ctx, configKey)
	} else if strings.HasPrefix(key, secret.ProtocolCredential) {
		key := key[len(secret.ProtocolCredential):]
		keys := strings.Split(key, ".")
		credential, err := v.getCredential(ctx, keys[0], "", "")
		if err != nil {
			v.logger.Error().Str("op", "Secret").Str("tid", v.tid.ToString()).Str("gears.app.id", v.tid.AppId).Str("partner.id", v.tid.OrgId).Msg("credential error: " + err.Error())
			return key
		}
		if len(keys) >= 2 {
			if strings.ToLower(keys[1]) == "id" {
				_, id := credential.GetId()
				return id
			}
			if strings.ToLower(keys[1]) == "scope" {
				_, scope := credential.GetScope()
				return scope
			}
			if strings.ToLower(keys[1]) == "granttype" {
				_, granttype := credential.GetGrantType()
				return granttype
			}
			if strings.ToLower(keys[1]) == "issuer" {
				_, issuer := credential.GetIssuer()
				return issuer
			}
			if strings.ToLower(keys[1]) == "secret" {
				_, secret := credential.GetSecret()
				return secret
			}
		}
		_, secret := credential.GetSecret()
		return secret
	} else {
		return key
	}
}
