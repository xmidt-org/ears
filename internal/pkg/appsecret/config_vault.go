package appsecret

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

//TODO: ensure credentials api support is optional
//TODO: blank out client secret in getTenantAPI
//TODO: get credentials
//TODO: unit tests
//TODO: support sat in http plugin
//TODO: return key instead of blank when not a secret (adjust usage)
//TODO: treat every config as potentially dynamic evaluation or secret
//TODO: manage http client better

//ConfigVault provides secrets from ears app configuration
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

const SAT_URL = "https://sat-prod.codebig2.net/oauth/token"
const CREDENTIAL_URL = "https://{{env}}.gears.comcast.com/v2/applications/{{app}}/credentials/{{key}}"

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
		return ""
	} else {
		return ""
	}
}

type TenantConfigVault struct {
	parentVault  secret.Vault
	tid          tenant.Id
	tenantStorer tenant.TenantStorer
	httpClient   *http.Client
}

func NewTenantConfigVault(tid tenant.Id, parentVault secret.Vault, tenantStorer tenant.TenantStorer) secret.Vault {
	tcv := &TenantConfigVault{
		parentVault:  parentVault,
		tid:          tid,
		tenantStorer: tenantStorer,
	}
	return tcv
}

func (v *TenantConfigVault) getSatBearerToken(ctx context.Context) string {
	//curl -s -X POST -H "X-Client-Id: ***" -H "X-Client-Secret: ***" -H "Cache-Control: no-cache" https://sat-prod.codebig2.net/oauth/token
	//echo "Bearer $TOKEN"
	if time.Now().Unix() >= satToken.ExpiresAt {
		satToken = SatToken{}
	}
	if satToken.AccessToken != "" {
		return satToken.TokenType + " " + satToken.AccessToken
	}
	if v.httpClient == nil {
		v.httpClient = &http.Client{
			Timeout: 10 * time.Second,
		}
	}
	req, err := http.NewRequest("POST", SAT_URL, nil)
	if err != nil {
		return ""
	}
	tenantConfig, err := v.tenantStorer.GetConfig(ctx, v.tid)
	if err != nil {
		return ""
	}
	if len(tenantConfig.ClientIds) == 0 {
		return ""
	}
	if tenantConfig.ClientSecret == "" {
		return ""
	}
	req.Header.Add("X-Client-Id", tenantConfig.ClientIds[0])
	req.Header.Add("X-Client-Secret", tenantConfig.ClientSecret)
	req.Header.Add("Cache-Control", "no-cache")
	resp, err := v.httpClient.Do(req)
	if err != nil {
		return ""
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	err = json.Unmarshal(buf, &satToken)
	if err != nil {
		return ""
	}
	satToken.ExpiresAt = time.Now().Unix() + int64(satToken.ExpiresIn)
	return satToken.TokenType + " " + satToken.AccessToken
}

func (v *TenantConfigVault) getCredential(ctx context.Context, key string, credentialType string, field string) (*Credential, error) {
	token := v.getSatBearerToken(ctx)
	if token == "" {
		return nil, errors.New("no bearer token")
	}
	if v.httpClient == nil {
		v.httpClient = &http.Client{
			Timeout: 10 * time.Second,
		}
	}
	url := CREDENTIAL_URL
	url = strings.Replace(url, "{{env}}", "qa", -1)
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

/*func populateCred(item Cred, cred *Credential) {
	cred.Name = item.Name
	cred.Type = item.Type
	cred.Ttl = item.Ttl
	switch cred.Type {
	case "autobahn":
		cred.Autobahn = item.Data.Autobahn
	case "basic":
		cred.Basic = item.Data.Basic
	case "generic":
		cred.Basic = item.Data.Generic
	case "ssh":
		cred.Basic = item.Data.SSH
	case "oauth1":
		cred.Oauth = item.Data.OAuth1
	case "oauth2":
		cred.Oauth = item.Data.OAuth2
	case "sat":
		cred.Oauth = item.Data.SAT
	}
}*/

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
		credential, err := v.getCredential(ctx, key[len(secret.ProtocolCredential):], "", "")
		if err != nil {
			return ""
		}
		return credential.Data.Generic.Secret
	} else {
		return ""
	}
}
