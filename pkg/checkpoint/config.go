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
			defaultStorageConfig.StorageType = storageType
		}
		region := config.GetString("ears.checkpointer.region")
		if region != "" {
			defaultStorageConfig.Region = region
		}
		table := config.GetString("ears.checkpointer.table")
		if table != "" {
			defaultStorageConfig.Table = table
		}
		updateFrequencySeconds := config.GetInt("ears.checkpointer.updateFrequencySeconds")
		if updateFrequencySeconds > 0 {
			defaultStorageConfig.UpdateFrequencySeconds = updateFrequencySeconds
		}
		env := config.GetString("ears.env")
		if env != "" {
			defaultStorageConfig.Env = env
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
