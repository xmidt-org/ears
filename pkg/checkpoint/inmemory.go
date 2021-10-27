package checkpoint

import (
	"errors"
	"time"
)

func newInMemoryCheckpointManager(config StorageConfig) CheckpointManager {
	cm := new(InMemoryCheckpointManager)
	cm.StorageConfig = config
	return cm
}

func (cm *InMemoryCheckpointManager) GetCheckpoint(Id string) (string, time.Time, error) {
	cm.Lock()
	defer cm.Unlock()
	return cm.checkpoints[Id], time.Now(), nil
}

func (cm *InMemoryCheckpointManager) SetCheckpoint(Id string, sequenceNumber string) error {
	cm.Lock()
	defer cm.Unlock()
	if sequenceNumber == "" {
		return errors.New("cannot pass blank sequence number as checkpoint")
	}
	cm.checkpoints[Id] = sequenceNumber
	return nil
}

func (cm *InMemoryCheckpointManager) GetId(pluginHash, consumerName, streamName, shardID string) string {
	return cm.Env + "-" + pluginHash + "-" + consumerName + "-" + streamName + "-" + shardID
}
