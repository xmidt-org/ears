package sharder

import (
	"github.com/xmidt-org/ears/internal/pkg/config"
	"net"
	"os"
)

var (
	DefaultDistributorConfigs = map[string]string{
		"type": "dynamodb",
		//"region": "local",
		"region":          "us-west-2",
		"table":           "ears-nodes",
		"updateFrequency": "10",
		"olderThan":       "60",
		"tag":             "dev",
	}
	defaultNodeStateManager NodeStateManager
)

func InitDistributorConfigs(config config.Config) {
	storageType := config.GetString("ears.sharder.type")
	if storageType != "" {
		DefaultDistributorConfigs["type"] = storageType
	}
	region := config.GetString("ears.sharder.region")
	if region != "" {
		DefaultDistributorConfigs["region"] = region
	}
	table := config.GetString("ears.sharder.table")
	if table != "" {
		DefaultDistributorConfigs["table"] = table
	}
	updateFrequency := config.GetString("ears.sharder.updateFrequency")
	if updateFrequency != "" {
		DefaultDistributorConfigs["updateFrequency"] = updateFrequency
	}
	olderThan := config.GetString("ears.sharder.olderThan")
	if olderThan != "" {
		DefaultDistributorConfigs["olderThan"] = olderThan
	}
	tag := config.GetString("ears.env")
	if tag != "" {
		DefaultDistributorConfigs["tag"] = tag
	}
}

// DefaultControllerConfig generates a default configuration based on a local dynamo instance
func DefaultControllerConfig(config config.Config) *ControllerConfig {
	cc := ControllerConfig{
		ShardConfig: ShardConfig{
			NumShards:   0,
			OwnedShards: []string{},
		},
		StorageConfig: DefaultDistributorConfigs,
	}
	if config != nil {
		customType := config.GetString("ears.sharder.type")
		if customType != "" {
			cc.StorageConfig["type"] = customType
		}
		customRegion := config.GetString("ears.sharder.region")
		if customRegion != "" {
			cc.StorageConfig["region"] = customRegion
		}
		customTable := config.GetString("ears.sharder.table")
		if customTable != "" {
			cc.StorageConfig["table"] = customTable
		}
		customUpdateFrequency := config.GetString("ears.sharder.updateFrequency")
		if customUpdateFrequency != "" {
			cc.StorageConfig["updateFrequency"] = customUpdateFrequency
		}
		customOlderThan := config.GetString("ears.sharder.olderThan")
		if customOlderThan != "" {
			cc.StorageConfig["olderThan"] = customOlderThan
		}
		customTag := config.GetString("ears.env")
		if customTag != "" {
			cc.StorageConfig["tag"] = customTag
		}
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

func GetDefaultNodeStateManager(identity string, configData map[string]string) (NodeStateManager, error) {
	var err error
	if defaultNodeStateManager == nil {
		if configData["table"] == "dynamodb" {
			defaultNodeStateManager, err = newDynamoDBNodeManager(identity, configData)
		} else if configData["table"] == "inmemory" {
			defaultNodeStateManager, err = newInMemoryNodeManager(identity, configData)
		}
	}
	return defaultNodeStateManager, err
}

func GetDefaultHashDistributor(identity string, numShards int, configData map[string]string) (ShardDistributor, error) {
	return newSimpleHashDistributor(identity, numShards, configData)
}
