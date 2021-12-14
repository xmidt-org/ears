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
