package appsecret

import (
	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
	"strings"
)

//ConfigVault provides secrets from ears app configuration
type ConfigVault struct {
	config config.Config
}

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
	parentVault secret.Vault
	tid         tenant.Id
}

func NewTenantConfigVault(tid tenant.Id, parentVault secret.Vault) secret.Vault {
	return &TenantConfigVault{
		parentVault: parentVault,
		tid:         tid,
	}
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
