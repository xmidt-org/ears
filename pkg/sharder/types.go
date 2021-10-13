package sharder

import (
	"strings"
	"sync"
	"time"
)

// Timestamp formats into the one true time format
func Timestamp(t time.Time) string {
	ts := t.UTC().Format(time.RFC3339Nano)
	//Format library cuts off trailing 0s. See: https://github.com/golang/go/issues/19635
	//This cause probably if we want to compare natural ordering of the string.
	//Artificially add the precision back
	if !strings.Contains(ts, ".") {
		ts = ts[0:len(ts)-1] + ".000000000Z"
	}
	return ts
}

func ParseTime(timestamp string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, timestamp)
}

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
