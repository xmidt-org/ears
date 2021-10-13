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
	DefaultClusterName = "dev"
	// DefaultDistributor is the default shard distribution mechanism
	DefaultDistributor = "manual"
)

var (
	DefaultDistributorStorageMap = map[string]string{
		"type": "dynamo",
		//"region": "local",
		"region":          "us-west-2",
		"table":           "ears-peers",
		"updateFrequency": "10",
		"olderThan":       "60",
		"tag":             "dev",
	}
)

func InitDistributorStorageMap(config config.Config) {
	storageType := config.GetString("ears.sharder.type")
	if storageType != "" {
		DefaultDistributorStorageMap["type"] = storageType
	}
	region := config.GetString("ears.sharder.region")
	if region != "" {
		DefaultDistributorStorageMap["region"] = region
	}
	table := config.GetString("ears.sharder.table")
	if table != "" {
		DefaultDistributorStorageMap["table"] = table
	}
	updateFrequency := config.GetString("ears.sharder.updateFrequency")
	if updateFrequency != "" {
		DefaultDistributorStorageMap["updateFrequency"] = updateFrequency
	}
	olderThan := config.GetString("ears.sharder.olderThan")
	if olderThan != "" {
		DefaultDistributorStorageMap["olderThan"] = olderThan
	}
	tag := config.GetString("ears.env")
	if tag != "" {
		DefaultDistributorStorageMap["tag"] = tag
	}
}

// DefaultControllerConfig generates a default configuration based on a local dynamo instance
func DefaultControllerConfig() *ControllerConfig {
	cc := ControllerConfig{
		Name:  DefaultClusterName,
		Local: true,
		ShardConfig: ShardConfig{
			NumShards:   DefaultNumShards,
			OwnedShards: []string{"0"},
		},
		Storage:     DefaultDistributorStorageMap,
		Distributor: DefaultDistributor,
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
