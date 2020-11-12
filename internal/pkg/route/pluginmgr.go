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

package route

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"
)

type (
	DefaultIOPluginManager struct {
		pluginMap map[string]Pluginer
		lock      sync.RWMutex
	}
)

var (
	pluginMgr *DefaultIOPluginManager
)

func NewIOPluginManager(ctx context.Context) *DefaultIOPluginManager {
	pm := new(DefaultIOPluginManager)
	pm.pluginMap = make(map[string]Pluginer)
	return pm
}

func GetIOPluginManager(ctx context.Context) *DefaultIOPluginManager {
	if pluginMgr == nil {
		pluginMgr = NewIOPluginManager(ctx)
	}
	return pluginMgr
}

func (pm *DefaultIOPluginManager) String() string {
	buf, _ := json.MarshalIndent(pm.pluginMap, "", "\t")
	return string(buf)
}

func (pm *DefaultIOPluginManager) RegisterPlugin(ctx context.Context, rte *Route, plugin Pluginer) (Pluginer, error) {
	pm.lock.Lock()
	defer pm.lock.Unlock()
	var err error
	hash := plugin.Hash(ctx)
	if hash == "" {
		return plugin, new(EmptyPluginHashError)
	}
	if p, ok := pm.pluginMap[hash]; ok {
		pc := p.GetConfig()
		if pc.routes == nil {
			pc.routes = make([]*Route, 0)
		}
		pc.routes = append(pc.routes, rte)
		log.Ctx(ctx).Debug().Msg(fmt.Sprintf("appending route to plugin %s %s route count %d %d", p.GetConfig().Name, hash, p.GetRouteCount(), len(pc.routes)))
		return p, nil
	}
	var p Pluginer
	if plugin.GetConfig().Mode == PluginModeInput {
		p, err = NewInputPlugin(ctx, rte)
	} else if plugin.GetConfig().Mode == PluginModeOutput {
		p, err = NewOutputPlugin(ctx, rte)
	}
	if err != nil {
		return plugin, err
	}
	pm.pluginMap[hash] = p
	log.Ctx(ctx).Debug().Msg(fmt.Sprintf("registering new %s %s plugin %s with hash %s", p.GetConfig().Type, p.GetConfig().Mode, p.GetConfig().Name, hash))
	return p, nil
}

func (pm *DefaultIOPluginManager) WithdrawPlugin(ctx context.Context, rte *Route, plugin Pluginer) error {
	pm.lock.Lock()
	defer pm.lock.Unlock()
	hash := plugin.Hash(ctx)
	if hash == "" {
		return new(EmptyPluginHashError)
	}
	if p, ok := pm.pluginMap[hash]; ok {
		pc := p.GetConfig()
		log.Ctx(ctx).Debug().Msg(fmt.Sprintf("withdrawing %s %s plugin %s with hash %s", p.GetConfig().Type, p.GetConfig().Mode, p.GetConfig().Name, hash))
		routes := make([]*Route, 0)
		if pc.routes != nil {
			for _, r := range pc.routes {
				if r.Hash(ctx) != rte.Hash(ctx) {
					routes = append(routes, r)
				}
			}
			pc.routes = routes
			if p.GetRouteCount() <= 0 {
				p.Close(ctx)
				log.Ctx(ctx).Debug().Msg(fmt.Sprintf("stopped %s %s plugin %s with hash %s", p.GetConfig().Type, p.GetConfig().Mode, p.GetConfig().Name, hash))
				delete(pm.pluginMap, hash)
			}
		}
		log.Ctx(ctx).Debug().Msg(fmt.Sprintf("%s %s plugin %s route count %d %d after withdrawal", p.GetConfig().Type, p.GetConfig().Mode, p.GetConfig().Name, p.GetRouteCount(), len(pc.routes)))
	}
	return nil
}

func (pm *DefaultIOPluginManager) GetPluginCount(ctx context.Context) int {
	return len(pm.pluginMap)
}
func (pm *DefaultIOPluginManager) GetAllPlugins(ctx context.Context) ([]Pluginer, error) {
	plugins := make([]Pluginer, 0)
	for _, p := range pm.pluginMap {
		plugins = append(plugins, p)
	}
	return plugins, nil
}

func (pm *DefaultIOPluginManager) GetPlugins(ctx context.Context, pluginMode string, pluginType string) ([]Pluginer, error) {
	plugins := make([]Pluginer, 0)
	for _, p := range pm.pluginMap {
		if pluginMode != "" && p.GetConfig().Mode != pluginMode {
			continue
		}
		if pluginType != "" && p.GetConfig().Type != pluginType {
			continue
		}
		plugins = append(plugins, p)
	}
	return plugins, nil
}
