package secret

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

func (v *ConfigVault) Secret(tid tenant.Id, key string) string {
	configKey := "ears.secrets." + tid.OrgId + "." + tid.AppId + "." + key
	return v.config.GetString(configKey)
}
