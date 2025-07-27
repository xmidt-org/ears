// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package nodestatemanagerfx

import (
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/pkg/sharder"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		ProvideNoteStateManager,
	),
)

type NodeStateManagerIn struct {
	fx.In
	Config config.Config
	Logger *zerolog.Logger
}

type NodeStateManagerOut struct {
	fx.Out
	NodeStateManager sharder.NodeStateManager
}

func ProvideNoteStateManager(in NodeStateManagerIn) (NodeStateManagerOut, error) {
	out := NodeStateManagerOut{}
	sharder.InitDistributorConfigs(in.Config)
	sharderConfig := sharder.DefaultControllerConfig()
	var err error
	out.NodeStateManager, err = sharder.GetDefaultNodeStateManager(sharderConfig.Identity, sharderConfig.StorageConfig)
	return out, err
}
