package sharder

import (
	"net"
	"os"
	"sync"
)

const (
	// DefaultNumShards is the default number of shards for the cluster as a whole
	DefaultNumShards = 1
	// DefaultClusterName is the default name of the cluster
	DefaultClusterName = "dev"
	// DefaultDistributor is the default shard distribution mechanism
	DefaultDistributor = "manual"
)

// Controller represents persistence and work sharing elements
type Controller struct {
	Distributor ShardDistributor
	ShardConfig
	StorageConfig map[string]string
	sync.RWMutex
}

// ShardStatus represents a single shard's status
type ShardStatus struct {
	Shard      string `json:"shard"`
	LastPolled string `json:"lastPolledAt"`
}

// ShardConfig represents the shard ownership of each node
type ShardConfig struct {
	IP          string   `json:"IP"`
	NumShards   int      `json:"clusterShardNumber"`
	OwnedShards []string `json:"ownedShards"`
}

// ShardUpdater will return a ShardConfig whenever the configuration changes
type ShardUpdater chan ShardConfig

// ShardDistributor defines
type ShardDistributor interface {
	// Stop shuts down any resources used
	Stop()
	// Status returns whether the ShardDistributor is healthy
	//Status() error
	// Updates
	//return the ShardConfig channel
	Updates() ShardUpdater
	// Peers returns the list of healthy peers
	Peers() []string
	// Identity returns the identity of this node
	Identity() string
}

// ControllerConfig contains cluster and node based configuration data
type ControllerConfig struct {
	Name  string
	Local bool
	ShardConfig
	Storage     map[string]string
	NodeName    string
	Distributor string
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
		Storage: map[string]string{"type": "dynamo",
			"region": "local",
			"table":  "ears-peers",
		},
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
