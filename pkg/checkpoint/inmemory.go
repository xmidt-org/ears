package checkpoint

import "errors"

func newInMemoryCheckpointManager(config StorageConfig) CheckpointManager {
	cm := new(InMemoryCheckpointManager)
	cm.StorageConfig = config
	return cm
}

func (cm *InMemoryCheckpointManager) GetCheckpoint(Id string) (string, string, error) {
	cm.Lock()
	defer cm.Unlock()
	return cm.checkpoints[Id], "", nil
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
