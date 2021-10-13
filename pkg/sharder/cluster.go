package sharder

import (
	"github.com/xmidt-org/ears/internal/pkg/config"
	"net"
	"os"
)

const (
	// DefaultNumShards is the default number of shards for the cluster as a whole
	DefaultNumShards = 1
	// DefaultClusterName is the default name of the cluster
	//DefaultClusterName = "ears"
)

var (
	DefaultDistributorConfigs = map[string]string{
		"type": "dynamo",
		//"region": "local",
		"region":          "us-west-2",
		"table":           "ears-nodes",
		"updateFrequency": "10",
		"olderThan":       "60",
		"tag":             "dev",
	}
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
func DefaultControllerConfig() *ControllerConfig {
	cc := ControllerConfig{
		//Name: DefaultClusterName,
		ShardConfig: ShardConfig{
			NumShards:   DefaultNumShards,
			OwnedShards: []string{"0"},
		},
		StorageConfig: DefaultDistributorConfigs,
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
					break
				}
			}
		}
	}
	return &cc
}
