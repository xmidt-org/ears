package sharder

import (
	"github.com/rs/zerolog"
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
	Identity    string   `json:"identity"`
	NumShards   int      `json:"clusterShardNumber"`
	OwnedShards []string `json:"ownedShards"`
}

// ShardUpdater will return a ShardConfig whenever the configuration changes
type ShardUpdater chan ShardConfig

// interface to get node state in the cluster
type NodeStateManager interface {
	GetActiveNodes() ([]string, error)
	RemoveNode()
}

// ShardDistributor defines
type ShardDistributor interface {
	// Stop shuts down any resources used
	Stop()
	// Updates
	//return the ShardConfig channel
	Updates() ShardUpdater
	// Peers returns the list of healthy nodes
	Nodes() []string
	// Identity returns the identity of this node
	Identity() string
	// UpdateNumberShards sets a new number of shards, rehashes shards and publishes updates
	UpdateNumberShards(numShards int)
}

// ControllerConfig contains cluster and node based configuration data
type ControllerConfig struct {
	//Name string
	ShardConfig
	StorageConfig map[string]string
	NodeName      string
}

type SimpleHashDistributor struct {
	ShardConfig
	nodeManager NodeStateManager
	identity    string
	nodes       []string
	updateChan  chan ShardConfig
	logger      *zerolog.Logger
	sync.Mutex
}
