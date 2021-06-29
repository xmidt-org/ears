package appsecret_test

import (
	"github.com/sebdah/goldie/v2"
	"github.com/spf13/viper"
	"github.com/xmidt-org/ears/internal/pkg/appsecret"
	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/pkg/tenant"
	"testing"
)

func SetupConfig(t *testing.T) config.Config {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	err := v.ReadInConfig()
	if err != nil {
		t.Fatalf("Fail to load test configuration %s\n", err.Error())
	}
	return v
}

func TestConfigVault(t *testing.T) {
	config := SetupConfig(t)
	v := appsecret.NewConfigVault(config)

	g := goldie.New(t)
	val := v.Secret("secret://myorg.myapp.kafka.secret1")
	g.Assert(t, "secret1", []byte(val))
	val = v.Secret("secret://myorg.myapp.kafka.secret2")
	g.Assert(t, "secret2", []byte(val))

	val = v.Secret("myorg.myapp.kafka.secret1")
	if val != "" {
		t.Fatalf("Expect empty secret, got %s\n", val)
	}

	v = appsecret.NewTenantConfigVault(tenant.Id{OrgId: "myorg", AppId: "myapp"}, v)
	val = v.Secret("secret://kafka.secret1")
	g.Assert(t, "secret1", []byte(val))
	val = v.Secret("secret://kafka.secret2")
	g.Assert(t, "secret2", []byte(val))

	val = v.Secret("kafka.secret1")
	if val != "" {
		t.Fatalf("Expect empty secret, got %s\n", val)
	}
}
