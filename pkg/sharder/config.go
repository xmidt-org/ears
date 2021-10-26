package sharder

import (
	"errors"
	"github.com/xmidt-org/ears/internal/pkg/config"
	"net"
	"os"
)

const (
	storageTypeDynamo             = "dynamodb"
	defaultUpdateFrequencySeconds = 10
	defaultUpdateTtlSeconds       = 60
	defaultTable                  = "ears-nodes"
	defaultTag                    = "local"
	defaultRegion                 = "us-west-2"
	//defaultRegion          = "local"
)

type (
	SharderConfig struct {
		StorageType            string
		StorageRegion          string
		StorageTable           string
		StorageTag             string
		UpdateFrequencySeconds int
		UpdateTtlSeconds       int
	}
)

var (
	DefaultSharderConfig = SharderConfig{
		StorageType:            storageTypeDynamo,
		StorageRegion:          defaultRegion,
		StorageTable:           defaultTable,
		StorageTag:             defaultTag,
		UpdateFrequencySeconds: defaultUpdateFrequencySeconds,
		UpdateTtlSeconds:       defaultUpdateTtlSeconds,
	}
	defaultNodeStateManager NodeStateManager
)

func InitDistributorConfigs(config config.Config) {
	storageType := config.GetString("ears.sharder.type")
	if storageType != "" {
		DefaultSharderConfig.StorageType = storageType
	}
	region := config.GetString("ears.sharder.region")
	if region != "" {
		DefaultSharderConfig.StorageRegion = region
	}
	table := config.GetString("ears.sharder.table")
	if table != "" {
		DefaultSharderConfig.StorageTable = table
	}
	updateFrequencySeconds := config.GetInt("ears.sharder.updateFrequencySeconds")
	if updateFrequencySeconds > 0 {
		DefaultSharderConfig.UpdateFrequencySeconds = updateFrequencySeconds
	}
	updateTtlSeconds := config.GetInt("ears.sharder.updateTtlSeconds")
	if updateTtlSeconds > 0 {
		DefaultSharderConfig.UpdateTtlSeconds = updateTtlSeconds
	}
	tag := config.GetString("ears.env")
	if tag != "" {
		DefaultSharderConfig.StorageTag = tag
	}
}

// DefaultControllerConfig generates a default configuration based on a local dynamo instance
func DefaultControllerConfig() *ControllerConfig {
	cc := ControllerConfig{
		ShardConfig: ShardConfig{
			NumShards:   0,
			OwnedShards: []string{},
		},
		StorageConfig: DefaultSharderConfig,
	}
	cc.NodeName = os.Getenv("HOSTNAME")
	if "" == cc.NodeName {
		address, err := net.InterfaceAddrs()
		if err != nil {
			panic(err)
		}
		for _, addr := range address {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					cc.NodeName = ipnet.IP.String()
					cc.Identity = cc.NodeName
					break
				}
			}
		}
	}
	return &cc
}

func GetDefaultNodeStateManager(identity string, configData SharderConfig) (NodeStateManager, error) {
	var err error
	if defaultNodeStateManager == nil {
		if configData.StorageTable == "dynamodb" {
			defaultNodeStateManager, err = newDynamoDBNodeManager(identity, configData)
		} else if configData.StorageTable == "inmemory" {
			defaultNodeStateManager, err = newInMemoryNodeManager(identity, configData)
		} else {
			return nil, errors.New("unknown sharder storage type " + configData.StorageTable)
		}
	}
	return defaultNodeStateManager, err
}

func GetDefaultHashDistributor(identity string, numShards int, configData SharderConfig) (ShardDistributor, error) {
	return newSimpleHashDistributor(identity, numShards, configData)
}
