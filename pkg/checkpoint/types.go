// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package checkpoint

import (
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type (
	CheckpointManager interface {
		GetCheckpoint(Id string) (string, time.Time, error)
		SetCheckpoint(Id string, sequenceNumber string) error
	}

	DynamoCheckpointManager struct {
		sync.Mutex
		StorageConfig
		lastUpdate map[string]time.Time
		svc        *dynamodb.DynamoDB
	}

	InMemoryCheckpointManager struct {
		sync.Mutex
		StorageConfig
		checkpoints map[string]string
	}

	StorageConfig struct {
		Env                    string
		StorageType            string
		Table                  string
		Region                 string
		UpdateFrequencySeconds int
	}
)
