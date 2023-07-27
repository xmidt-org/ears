package appsecret

import (
	"encoding/json"
	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

//ConfigVault provides secrets from ears app configuration
type ConfigVault struct {
	config config.Config
}

type SatToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
	ExpiresAt   int64  `json:"timestamp"`
}

const SAT_URL = "https://sat-prod.codebig2.net/oauth/token"

var satToken SatToken

func NewConfigVault(config config.Config) secret.Vault {
	return &ConfigVault{config: config}
}

func (v *ConfigVault) Secret(key string) string {
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
	tenantStorer *tenant.TenantStorer
}

func NewTenantConfigVault(tid tenant.Id, parentVault secret.Vault, tenantStorer *tenant.TenantStorer) secret.Vault {
	tcv := &TenantConfigVault{
		parentVault:  parentVault,
		tid:          tid,
		tenantStorer: tenantStorer,
	}
	return tcv
}

func (v *TenantConfigVault) getSatBearerToken() string {
	//curl -s -X POST -H "X-Client-Id: ***" -H "X-Client-Secret: ***" -H "Cache-Control: no-cache" https://sat-prod.codebig2.net/oauth/token
	//echo "Bearer $TOKEN"
	if time.Now().Unix() >= satToken.ExpiresAt {
		satToken = SatToken{}
	}
	if satToken.AccessToken != "" {
		return satToken.TokenType + " " + satToken.AccessToken
	}
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("POST", SAT_URL, nil)
	if err != nil {
		return ""
	}
	tenantConfig, err := (*v.tenantStorer).GetConfig(nil, v.tid)
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
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	err = json.Unmarshal(buf, &satToken)
	if err != nil {
		return ""
	}
	satToken.ExpiresAt = time.Now().Unix() + int64(satToken.ExpiresIn)
	return satToken.TokenType + " " + satToken.AccessToken
}

func (v *TenantConfigVault) Secret(key string) string {
	if strings.HasPrefix(key, secret.ProtocolSecret) {
		configKey := secret.ProtocolSecret + v.tid.OrgId + "." + v.tid.AppId + "." + key[len(secret.ProtocolSecret):]
		val := v.parentVault.Secret(configKey)
		if val != "" {
			return val
		}
		// try again with global key/secrets
		configKey = secret.ProtocolSecret + "all.all." + key[len(secret.ProtocolSecret):]
		return v.parentVault.Secret(configKey)
	} else if strings.HasPrefix(key, secret.ProtocolCredential) {
		//TODO: query credentials API here
		return ""
	} else {
		return ""
	}
}
