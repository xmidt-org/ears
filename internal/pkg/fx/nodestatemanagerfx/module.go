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
