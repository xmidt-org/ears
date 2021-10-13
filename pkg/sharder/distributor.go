package sharder

import (
	"errors"
	"math/rand"
	"sort"
	"strconv"
	"sync"
	"time"
)

type (
	SimpleHashDistributor struct {
		ShardConfig
		nodeManager NodeStateManager
		identity    string
		peers       []string
		updateChan  chan ShardConfig
		enabled     bool
		sync.Mutex
	}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func NewDynamoSimpleHashDistributor(identity string, numShards int, configData map[string]string) (*SimpleHashDistributor, error) {
	d := &SimpleHashDistributor{
		enabled:    true,
		identity:   identity,
		updateChan: make(ShardUpdater),
		ShardConfig: ShardConfig{
			NumShards: numShards,
			IP:        identity,
		},
	}
	healthTableName, found := configData["healthTable"]
	if !found {
		return nil, errors.New("healthTable must be set")
	}
	var updateFrequency, olderThan int
	if frq, found := configData["updateFrequency"]; found {
		updateFrequency, _ = strconv.Atoi(frq)
	}
	if old, found := configData["olderThan"]; found {
		olderThan, _ = strconv.Atoi(old)
	}
	region := configData["region"]
	tag := configData["tag"]
	m, err := newDynamoDBNodesManager(healthTableName, region, identity, updateFrequency, olderThan, tag)
	if err != nil {
		return nil, err
	}
	d.nodeManager = m
	go d.peerMonitor()
	return d, nil
}

// Stop releases any resources
func (c *SimpleHashDistributor) Stop() {
}

// Updates returns the channel that SimpleHashDistributor will send ShardConfig updates on
func (c *SimpleHashDistributor) Updates() ShardUpdater {
	return c.updateChan
}

// Identity returns the note IP that SimpleHashDistributor locate
func (c *SimpleHashDistributor) Identity() string {
	return c.identity
}

// peerMonitor watches for changes in the health service
func (c *SimpleHashDistributor) peerMonitor() {
	defer c.nodeManager.CleanUp()
	go func() {
		defer c.nodeManager.CleanUp()
		for {
			time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
			aliveNodes, err := c.nodeManager.GetNodeState()
			if err != nil || aliveNodes == nil {
				//need add retry logical here later
				/*time.Sleep(time.Duration(rand.Intn(10)) * time.Second)*/
				continue
			}
			sort.Strings(aliveNodes)
			var change bool
			c.Lock()
			if len(aliveNodes) != len(c.peers) {
				change = true
			} else {
				for i, peer := range aliveNodes {
					if peer != c.peers[i] {
						change = true
						break
					}
				}
			}
			if change {
				c.peers = aliveNodes
				c.Unlock()
				c.ownership()
				continue
			}
			c.Unlock()
		}
	}()
}

// Peers returns our list of peers
func (c *SimpleHashDistributor) Peers() []string {
	c.Lock()
	defer c.Unlock()
	return c.peers
}

// ownership receives events about peer changes and determines if shard ownership has changed
// c.Lock() is held by peerMonitor() before calling us
func (c *SimpleHashDistributor) ownership() {
	if len(c.peers) == 0 {
		c.ShardConfig.OwnedShards = nil
		c.updateChan <- c.ShardConfig
		return
	}
	changeFlag := c.getShardsByHash()
	if changeFlag {
		c.updateChan <- c.ShardConfig
		return
	}
}

func (c *SimpleHashDistributor) getShardsByHash() bool {
	var shards []string
	var identityIndex int
	for i, peer := range c.peers {
		if peer == c.identity {
			identityIndex = i
			break
		}
	}
	for j := 0; j < c.ShardConfig.NumShards; j++ {
		if (j % len(c.peers)) == identityIndex {
			shards = append(shards, strconv.Itoa(j))
		}
	}
	// check the len of the shards change
	if len(shards) != len(c.ShardConfig.OwnedShards) {
		c.ShardConfig.OwnedShards = shards
		return true
	}
	// check the content of owned shard change
	for i, shard := range shards {
		if shard != c.ShardConfig.OwnedShards[i] {
			c.ShardConfig.OwnedShards = shards
			return true
		}
	}
	return false
}
