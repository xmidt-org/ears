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

package checkpointmanagerfx

import (
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/pkg/checkpoint"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		ProvideCheckpointManager,
	),
)

type CheckpointManagerIn struct {
	fx.In
	Config config.Config
	Logger *zerolog.Logger
}

type CheckpointManagerOut struct {
	fx.Out
	CheckpointManager checkpoint.CheckpointManager
}

func ProvideCheckpointManager(in CheckpointManagerIn) (CheckpointManagerOut, error) {
	out := CheckpointManagerOut{}
	var err error
	out.CheckpointManager, err = checkpoint.GetDefaultCheckpointManager(in.Config)
	return out, err
}
