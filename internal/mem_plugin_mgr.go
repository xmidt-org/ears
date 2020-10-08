/**
 *  Copyright (c) 2020  Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package internal

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
)

//TODO: merge input plugin mgr and output plugin mgr

type (
	DefaultInputPluginManager struct {
		pluginMap map[string]*InputPlugin
	}
	DefaultOutputPluginManager struct {
		pluginMap map[string]*OutputPlugin
	}
)

var (
	inputPluginMgr  InputPluginManager
	outputPluginMgr OutputPluginManager
)

func NewOutputPluginManager(ctx context.Context) OutputPluginManager {
	pm := new(DefaultOutputPluginManager)
	pm.pluginMap = make(map[string]*OutputPlugin)
	return pm
}

func GetOutputPluginManager(ctx context.Context) OutputPluginManager {
	if outputPluginMgr == nil {
		outputPluginMgr = NewOutputPluginManager(ctx)
	}
	return outputPluginMgr
}

func NewInputPluginManager(ctx context.Context) InputPluginManager {
	pm := new(DefaultInputPluginManager)
	pm.pluginMap = make(map[string]*InputPlugin)
	return pm
}

func GetInputPluginManager(ctx context.Context) InputPluginManager {
	if inputPluginMgr == nil {
		inputPluginMgr = NewInputPluginManager(ctx)
	}
	return inputPluginMgr
}

func (pm *DefaultInputPluginManager) RegisterRoute(ctx context.Context, rte *RoutingTableEntry) (*InputPlugin, error) {
	var err error
	plugin := rte.Source
	hash := plugin.Hash(ctx)
	if hash == "" {
		return plugin, errors.New("empty plugin hash")
	}
	if p, ok := pm.pluginMap[hash]; ok {
		p.RouteCount++
		if p.routes == nil {
			p.routes = make([]*RoutingTableEntry, 0)
		}
		p.routes = append(p.routes, rte)
		return p, nil
	}
	pm.pluginMap[hash], err = NewInputPlugin(ctx, rte)
	log.Debug().Msg(fmt.Sprintf("registering input plugin with hash %s", hash))
	if err != nil {
		return plugin, err
	}
	return pm.pluginMap[hash], nil
}

func (pm *DefaultOutputPluginManager) RegisterRoute(ctx context.Context, rte *RoutingTableEntry) (*OutputPlugin, error) {
	var err error
	plugin := rte.Destination
	hash := plugin.Hash(ctx)
	if hash == "" {
		return plugin, errors.New("empty plugin hash")
	}
	if p, ok := pm.pluginMap[hash]; ok {
		p.RouteCount++
		if p.routes == nil {
			p.routes = make([]*RoutingTableEntry, 0)
		}
		p.routes = append(p.routes, rte)
		return p, nil
	}
	pm.pluginMap[hash], err = NewOutputPlugin(ctx, rte)
	log.Debug().Msg(fmt.Sprintf("registering output plugin with hash %s", hash))
	if err != nil {
		return plugin, err
	}
	return pm.pluginMap[hash], nil
}

func (pm *DefaultInputPluginManager) UnregisterRoute(ctx context.Context, rte *RoutingTableEntry) error {
	plugin := rte.Source
	hash := plugin.Hash(ctx)
	if hash == "" {
		return errors.New("empty plugin hash")
	}
	if p, ok := pm.pluginMap[hash]; ok {
		log.Debug().Msg(fmt.Sprintf("unregistering input plugin with hash %s", hash))
		routes := make([]*RoutingTableEntry, 0)
		if p.routes != nil {
			for _, r := range p.routes {
				if r != rte {
					routes = append(routes, r)
				} else {
					p.RouteCount--
				}
			}
			if p.RouteCount <= 0 {
				//TODO: stop plugin
				log.Debug().Msg(fmt.Sprintf("stopped input plugin with hash %s", hash))
				delete(pm.pluginMap, hash)
			}
		}
		p.routes = routes
	}
	return nil
}

func (pm *DefaultOutputPluginManager) UnregisterRoute(ctx context.Context, rte *RoutingTableEntry) error {
	plugin := rte.Destination
	hash := plugin.Hash(ctx)
	if hash == "" {
		return errors.New("empty plugin hash")
	}
	if p, ok := pm.pluginMap[hash]; ok {
		log.Debug().Msg(fmt.Sprintf("unregistering output plugin with hash %s", hash))
		routes := make([]*RoutingTableEntry, 0)
		if p.routes != nil {
			for _, r := range p.routes {
				if r != rte {
					routes = append(routes, r)
				} else {
					p.RouteCount--
				}
			}
			if p.RouteCount <= 0 {
				//TODO: stop plugin
				log.Debug().Msg(fmt.Sprintf("stopped output plugin with hash %s", hash))
				delete(pm.pluginMap, hash)
			}
		}
		p.routes = routes
	}
	return nil
}
