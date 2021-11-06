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

package checkpoint

import (
	"errors"
	"github.com/xmidt-org/ears/internal/pkg/config"
)

const (
	defaultTableName              = "ears-checkpoints"
	defaultRegion                 = "us-west-2"
	defaultEnv                    = "local"
	defaultStorageType            = "inmemory"
	defaultUpdateFrequencySeconds = 60
)

var (
	defaultStorageConfig = StorageConfig{
		Table:                  defaultTableName,
		Region:                 defaultRegion,
		Env:                    defaultEnv,
		StorageType:            defaultStorageType,
		UpdateFrequencySeconds: defaultUpdateFrequencySeconds,
	}
	defaultCheckpointManager CheckpointManager
)

func GetStorageConfig(config config.Config) StorageConfig {
	storageConfig := defaultStorageConfig
	if config != nil {
		storageType := config.GetString("ears.checkpointer.type")
		if storageType != "" {
			storageConfig.StorageType = storageType
		}
		region := config.GetString("ears.checkpointer.region")
		if region != "" {
			storageConfig.Region = region
		}
		table := config.GetString("ears.checkpointer.table")
		if table != "" {
			storageConfig.Table = table
		}
		updateFrequencySeconds := config.GetInt("ears.checkpointer.updateFrequencySeconds")
		if updateFrequencySeconds > 0 {
			storageConfig.UpdateFrequencySeconds = updateFrequencySeconds
		}
		env := config.GetString("ears.env")
		if env != "" {
			storageConfig.Env = env
		}
	}
	return storageConfig
}

func GetDefaultCheckpointManager(config config.Config) (CheckpointManager, error) {
	if defaultCheckpointManager != nil {
		return defaultCheckpointManager, nil
	}
	var err error
	if config.GetString("ears.checkpointer.type") == "dynamodb" {
		defaultCheckpointManager, err = newDynamoCheckpointManager(GetStorageConfig(config))
	} else if config.GetString("ears.checkpointer.type") == "inmemory" {
		defaultCheckpointManager, err = newInMemoryCheckpointManager(GetStorageConfig(config)), nil
	} else {
		return nil, errors.New("unknown checkpoint manager storage type " + config.GetString("ears.checkpoint.type"))
	}
	return defaultCheckpointManager, err
}
