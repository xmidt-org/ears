// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package checkpoint

import (
	"errors"
	"time"
)

func newInMemoryCheckpointManager(config StorageConfig) CheckpointManager {
	cm := new(InMemoryCheckpointManager)
	cm.StorageConfig = config
	cm.checkpoints = map[string]string{}
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
