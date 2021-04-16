// Copyright 2020 Comcast Cable Communications Management, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package routestorerfx

import (
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/internal/pkg/db"
	"github.com/xmidt-org/ears/internal/pkg/db/bolt"
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
	storageType := in.Config.GetString("ears.storage.type")
	switch storageType {
	case "inmemory":
		out.RouteStorer = db.NewInMemoryRouteStorer(in.Config)
	case "dynamodb":
		routeStorer, err := dynamo.NewDynamoDbStorer(in.Config)
		if err != nil {
			return out, err
		}
		out.RouteStorer = routeStorer
	case "boltdb":
		routeStorer, err := bolt.NewBoltDbStorer(in.Config)
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
		return out, &UnsupportedStorageError{storageType}
	}
	return out, nil
}
