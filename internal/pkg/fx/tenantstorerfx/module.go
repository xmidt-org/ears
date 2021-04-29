package tenantstorerfx

import (
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/internal/pkg/db"
	"github.com/xmidt-org/ears/internal/pkg/db/dynamo"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		ProvideTenantStorer,
	),
)

type StorageIn struct {
	fx.In
	Config config.Config
	Logger *zerolog.Logger
}

type StorageOut struct {
	fx.Out
	TenantStorer tenant.TenantStorer
}

func ProvideTenantStorer(in StorageIn) (StorageOut, error) {
	out := StorageOut{}
	storageType := in.Config.GetString("ears.storage.route.type")
	switch storageType {
	case "inmemory":
		out.TenantStorer = db.NewTenantInmemoryStorer()
	case "dynamodb":
		tenantStorer, err := dynamo.NewTenantStorer(in.Config)
		if err != nil {
			return out, err
		}
		out.TenantStorer = tenantStorer
	default:
		return out, &UnsupportedTenantStorageError{storageType}
	}
	return out, nil
}
