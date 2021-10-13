package sharder

import (
	"sync"
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

// interface to get node state in the cluster
type NodeStateManager interface {
	GetClusterState() ([]string, error)
	RemoveNode()
}

// ShardDistributor defines
type ShardDistributor interface {
	// Stop shuts down any resources used
	Stop()
	// Status returns whether the ShardDistributor is healthy
	//Status() error
	// Updates
	//return the ShardConfig channel
	Updates() ShardUpdater
	// Peers returns the list of healthy nodes
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
