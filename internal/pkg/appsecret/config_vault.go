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
	if !strings.HasPrefix(key, secret.Protocol) {
		return ""
	}
	configKey := "ears.secrets." + key[len(secret.Protocol):]
	return v.config.GetString(configKey)
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
	if !strings.HasPrefix(key, secret.Protocol) {
		return ""
	}
	configKey := key[0:len(secret.Protocol)] + v.tid.OrgId + "." + v.tid.AppId + "." + key[len(secret.Protocol):]
	val := v.parentVault.Secret(configKey)
	if val != "" {
		return val
	}
	//try again with global key/secrets
	configKey = key[0:len(secret.Protocol)] + "all.all." + key[len(secret.Protocol):]
	return v.parentVault.Secret(configKey)
}
