// Copyright 2021 Comcast Cable Communications Management, LLC
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

package quotamanagerfx

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/internal/pkg/quota"
	"github.com/xmidt-org/ears/internal/pkg/syncer"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		ProvideQuotaManager,
	),
)

type QuotaManagerIn struct {
	fx.In
	Config       config.Config
	TenantStorer tenant.TenantStorer
	Syncer       syncer.DeltaSyncer
	Logger       *zerolog.Logger
}

type QuotaManagerOut struct {
	fx.Out
	QuotaManager *quota.QuotaManager
}

func ProvideQuotaManager(in QuotaManagerIn) (QuotaManagerOut, error) {
	out := QuotaManagerOut{}

	quotaMgr, err := quota.NewQuotaManager(in.Logger, in.TenantStorer, in.Syncer, in.Config)
	if err != nil {
		return out, err
	}
	out.QuotaManager = quotaMgr
	return out, nil
}

func SetupQuotaManager(lifecycle fx.Lifecycle, logger *zerolog.Logger, quotaManager *quota.QuotaManager) error {
	lifecycle.Append(
		fx.Hook{
			OnStart: func(context.Context) error {
				quotaManager.Start()
				logger.Info().Msg("Delta Syncer Service Started")
				return nil
			},
			OnStop: func(ctx context.Context) error {
				quotaManager.Stop()
				logger.Info().Msg("Delta Syncer Service Stopped")
				return nil
			},
		},
	)
	return nil
}
