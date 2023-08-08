package appsecret_test

import (
	"context"
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
	ctx := context.Background()
	val := v.Secret(ctx, "secret://myorg.myapp.kafka.secret1")
	g.Assert(t, "secret1", []byte(val))
	val = v.Secret(ctx, "secret://myorg.myapp.kafka.secret2")
	g.Assert(t, "secret2", []byte(val))

	val = v.Secret(ctx, "myorg.myapp.kafka.secret1")
	if val != "myorg.myapp.kafka.secret1" {
		t.Fatalf("expected non secret, got %s\n", val)
	}

	v = appsecret.NewTenantConfigVault(tenant.Id{OrgId: "myorg", AppId: "myapp"}, v, nil)
	val = v.Secret(ctx, "secret://kafka.secret1")
	g.Assert(t, "secret1", []byte(val))
	val = v.Secret(ctx, "secret://kafka.secret2")
	g.Assert(t, "secret2", []byte(val))

	val = v.Secret(ctx, "kafka.secret1")
	if val != "kafka.secret1" {
		t.Fatalf("expected non secret, got %s\n", val)
	}
}
