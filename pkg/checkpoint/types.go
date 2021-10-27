package checkpoint

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"sync"
	"time"
)

type (
	CheckpointManager interface {
		GetCheckpoint(Id string) (string, string, error)
		SetCheckpoint(Id string, sequenceNumber string) error
		GetId(pluginHash, consumerName, streamName, shardID string) string
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
