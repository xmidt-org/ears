// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package routestorerfx

import (
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/internal/pkg/db"
	"github.com/xmidt-org/ears/internal/pkg/db/dynamo"
	"github.com/xmidt-org/ears/internal/pkg/db/redis"
	"github.com/xmidt-org/ears/pkg/route"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		ProvideRouteStorer,
	),
)

type StorageIn struct {
	fx.In
	Config config.Config
	Logger *zerolog.Logger
}

type StorageOut struct {
	fx.Out
	RouteStorer route.RouteStorer
}

func ProvideRouteStorer(in StorageIn) (StorageOut, error) {
	out := StorageOut{}
	storageType := in.Config.GetString("ears.storage.route.type")
	switch storageType {
	case "inmemory":
		out.RouteStorer = db.NewInMemoryRouteStorer(in.Config)
	case "dynamodb":
		routeStorer, err := dynamo.NewDynamoDbStorer(in.Config)
		if err != nil {
			return out, err
		}
		out.RouteStorer = routeStorer
	case "redis":
		routeStorer, err := redis.NewRedisDbStorer(in.Config, in.Logger)
		if err != nil {
			return out, err
		}
		out.RouteStorer = routeStorer
	default:
		return out, &UnsupportedRouteStorageError{storageType}
	}
	return out, nil
}
