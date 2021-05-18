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

package syncerfx

import (
	"context"
	"errors"
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/internal/pkg/syncer"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		ProvideRouteTableSyncer,
	),
)

type TableSyncerIn struct {
	fx.In
	Config config.Config
	Logger *zerolog.Logger
}

type TableSyncerOut struct {
	fx.Out
	RoutingTableDeltaSyncer syncer.DeltaSyncer
}

func ProvideRouteTableSyncer(in TableSyncerIn) (TableSyncerOut, error) {
	out := TableSyncerOut{}
	tableSyncerType := in.Config.GetString("ears.synchronization.type")
	switch tableSyncerType {
	case "inmemory":
		out.RoutingTableDeltaSyncer = syncer.NewInMemoryDeltaSyncer(in.Logger, in.Config)
	case "redis":
		out.RoutingTableDeltaSyncer = syncer.NewRedisDeltaSyncer(in.Logger, in.Config)
	default:
		return out, errors.New("unsupported table syncer type " + tableSyncerType)
	}
	return out, nil
}

func SetupDeltaSyncer(lifecycle fx.Lifecycle, logger *zerolog.Logger, deltaSyncer syncer.DeltaSyncer) error {
	lifecycle.Append(
		fx.Hook{
			OnStart: func(context.Context) error {
				deltaSyncer.StartListeningForSyncRequests()
				logger.Info().Msg("Delta Syncer Service Started")
				return nil
			},
			OnStop: func(ctx context.Context) error {
				deltaSyncer.StopListeningForSyncRequests()
				logger.Info().Msg("Delta Syncer Service Stopped")
				return nil
			},
		},
	)
	return nil
}
