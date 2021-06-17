package appsecret

import (
	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
)

//ConfigVault provides secrets from ears app configuration
type ConfigVault struct {
	config config.Config
}

func NewConfigVault(config config.Config) secret.Vault {
	return &ConfigVault{config: config}
}

func (v *ConfigVault) Secret(key string) string {
	configKey := "ears.secrets." + key
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
	configKey := v.tid.OrgId + "." + v.tid.AppId + "." + key
	return v.parentVault.Secret(configKey)
}
