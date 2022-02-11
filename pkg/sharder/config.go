// Copyright 2020 Comcast Cable Communications Management, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	defaultForcedHostname         = ""
	//defaultRegion          = "local"
)

type (
	StorageConfig struct {
		StorageType            string
		StorageRegion          string
		StorageTable           string
		StorageTag             string
		UpdateFrequencySeconds int
		UpdateTtlSeconds       int
		ForcedHostname         string
	}
)

var (
	DefaultStorageConfig = StorageConfig{
		StorageType:            storageTypeDynamo,
		StorageRegion:          defaultRegion,
		StorageTable:           defaultTable,
		StorageTag:             defaultTag,
		UpdateFrequencySeconds: defaultUpdateFrequencySeconds,
		UpdateTtlSeconds:       defaultUpdateTtlSeconds,
		ForcedHostname:         defaultForcedHostname,
	}
	defaultNodeStateManager NodeStateManager
)

func InitDistributorConfigs(config config.Config) {
	storageType := config.GetString("ears.sharder.type")
	if storageType != "" {
		DefaultStorageConfig.StorageType = storageType
	}
	region := config.GetString("ears.sharder.region")
	if region != "" {
		DefaultStorageConfig.StorageRegion = region
	}
	table := config.GetString("ears.sharder.table")
	if table != "" {
		DefaultStorageConfig.StorageTable = table
	}
	forcedHostname := config.GetString("ears.hostname")
	if forcedHostname != "" {
		DefaultStorageConfig.ForcedHostname = forcedHostname
	}
	updateFrequencySeconds := config.GetInt("ears.sharder.updateFrequencySeconds")
	if updateFrequencySeconds > 0 {
		DefaultStorageConfig.UpdateFrequencySeconds = updateFrequencySeconds
	}
	updateTtlSeconds := config.GetInt("ears.sharder.updateTtlSeconds")
	if updateTtlSeconds > 0 {
		DefaultStorageConfig.UpdateTtlSeconds = updateTtlSeconds
	}
	tag := config.GetString("ears.env")
	if tag != "" {
		DefaultStorageConfig.StorageTag = tag
	}
}

// DefaultControllerConfig generates a default configuration based on a local dynamo instance
func DefaultControllerConfig() *ControllerConfig {
	cc := ControllerConfig{
		ShardConfig: ShardConfig{
			NumShards:   0,
			OwnedShards: []string{},
		},
		StorageConfig: DefaultStorageConfig,
	}
	cc.NodeName = cc.StorageConfig.ForcedHostname
	if cc.NodeName == "" {
		cc.NodeName, _ = os.Hostname()
		if cc.NodeName == "" {
			address, err := net.InterfaceAddrs()
			if err == nil {
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
		}
	}
	if cc.NodeName == "" {
		cc.NodeName = "unknown"
	}
	cc.ShardConfig.Identity = cc.NodeName
	return &cc
}

func GetDefaultNodeStateManager(identity string, configData StorageConfig) (NodeStateManager, error) {
	var err error
	if defaultNodeStateManager == nil {
		if configData.StorageType == "dynamodb" {
			defaultNodeStateManager, err = newDynamoDBNodeManager(identity, configData)
		} else if configData.StorageType == "inmemory" {
			defaultNodeStateManager, err = newInMemoryNodeManager(identity, configData)
		} else {
			return nil, errors.New("unknown sharder storage type " + configData.StorageTable)
		}
	}
	return defaultNodeStateManager, err
}

func GetDefaultHashDistributor(identity string, numShards int, configData StorageConfig) (ShardDistributor, error) {
	return newSimpleHashDistributor(identity, numShards, configData)
}
