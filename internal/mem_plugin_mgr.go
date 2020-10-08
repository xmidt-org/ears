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
		return plugin, errors.New("bad plugin configuration")
	}
	if p, ok := pm.pluginMap[hash]; ok {
		p.RouteCount++
		p.routes = append(p.routes, rte)
		return p, nil
	}
	pm.pluginMap[hash], err = NewInputPlugin(ctx, rte)
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
		return plugin, errors.New("bad plugin configuration")
	}
	if p, ok := pm.pluginMap[hash]; ok {
		p.RouteCount++
		p.routes = append(p.routes, rte)
		return p, nil
	}
	pm.pluginMap[hash], err = NewOutputPlugin(ctx, rte)
	if err != nil {
		return plugin, err
	}
	return pm.pluginMap[hash], nil
}

func (pm *DefaultInputPluginManager) UnregisterRoute(ctx context.Context, rte *RoutingTableEntry) error {
	plugin := rte.Source
	hash := plugin.Hash(ctx)
	if hash == "" {
		return errors.New("bad plugin configuration")
	}
	if p, ok := pm.pluginMap[hash]; ok {
		p.RouteCount--
		if p.RouteCount <= 0 {
			// stop plugin
		}
		routes := make([]*RoutingTableEntry, 0)
		for _, r := range p.routes {
			if r != rte {
				routes = append(routes, r)
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
		return errors.New("bad plugin configuration")
	}
	if p, ok := pm.pluginMap[hash]; ok {
		p.RouteCount--
		if p.RouteCount <= 0 {
			// stop plugin
		}
		routes := make([]*RoutingTableEntry, 0)
		for _, r := range p.routes {
			if r != rte {
				routes = append(routes, r)
			}
		}
		p.routes = routes
	}
	return nil
}
