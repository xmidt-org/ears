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
		nodes       []string
		updateChan  chan ShardConfig
		enabled     bool
		sync.Mutex
	}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func NewDynamoSimpleHashDistributor(identity string, numShards int, configData map[string]string) (*SimpleHashDistributor, error) {
	hashDistributor := &SimpleHashDistributor{
		enabled:    true,
		identity:   identity,
		updateChan: make(ShardUpdater),
		ShardConfig: ShardConfig{
			NumShards: numShards,
			Identity:  identity,
		},
	}
	tableName, found := configData["table"]
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
	nodeManager, err := newDynamoDBNodesManager(tableName, region, identity, updateFrequency, olderThan, tag)
	if err != nil {
		return nil, err
	}
	hashDistributor.nodeManager = nodeManager
	hashDistributor.nodeMonitor()
	return hashDistributor, nil
}

// Stop releases any resources
func (c *SimpleHashDistributor) Stop() {
}

// Updates returns the channel that SimpleHashDistributor will send ShardConfig updates on
func (c *SimpleHashDistributor) Updates() ShardUpdater {
	return c.updateChan
}

// Identity returns the note Identity that SimpleHashDistributor locate
func (c *SimpleHashDistributor) Identity() string {
	return c.identity
}

func (c *SimpleHashDistributor) UpdateNumberShards(numShards int) {
	c.NumShards = numShards
	c.publishChanges()
}

// nodeMonitor watches for changes in the health service
func (c *SimpleHashDistributor) nodeMonitor() {
	go func() {
		defer c.nodeManager.RemoveNode()
		for {
			time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
			aliveNodes, err := c.nodeManager.GetActiveNodes()
			if err != nil || aliveNodes == nil {
				continue
			}
			sort.Strings(aliveNodes)
			var change bool
			c.Lock()
			if len(aliveNodes) != len(c.nodes) {
				change = true
			} else {
				for i, peer := range aliveNodes {
					if peer != c.nodes[i] {
						change = true
						break
					}
				}
			}
			if change {
				c.nodes = aliveNodes
				c.Unlock()
				c.publishChanges()
				continue
			}
			c.Unlock()
		}
	}()
}

// Nodes returns list of nodes
func (c *SimpleHashDistributor) Nodes() []string {
	c.Lock()
	defer c.Unlock()
	return c.nodes
}

func (c *SimpleHashDistributor) publishChanges() {
	// c.Lock() is held by nodeMonitor() before calling us
	if len(c.nodes) == 0 {
		c.ShardConfig.OwnedShards = nil
		c.updateChan <- c.ShardConfig
		return
	}
	changeFlag := c.hashShards()
	if changeFlag {
		c.updateChan <- c.ShardConfig
		return
	}
}

func (c *SimpleHashDistributor) hashShards() bool {
	var myShards []string
	var myPeerIndex int
	for i, peer := range c.nodes {
		if peer == c.identity {
			myPeerIndex = i
			break
		}
	}
	for j := 0; j < c.ShardConfig.NumShards; j++ {
		if (j % len(c.nodes)) == myPeerIndex {
			myShards = append(myShards, strconv.Itoa(j))
		}
	}
	// check the len of the myShards change
	if len(myShards) != len(c.ShardConfig.OwnedShards) {
		c.ShardConfig.OwnedShards = myShards
		return true
	}
	// check the content of owned shard change
	for i, shard := range myShards {
		if shard != c.ShardConfig.OwnedShards[i] {
			c.ShardConfig.OwnedShards = myShards
			return true
		}
	}
	return false
}
