// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

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
