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
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
)

type (
	DefaultIOPluginManager struct {
		pluginMap map[string]Pluginer
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
	var err error
	hash := plugin.Hash(ctx)
	if hash == "" {
		return plugin, new(EmptyPluginHashError)
	}
	pc := plugin.GetConfig()
	if p, ok := pm.pluginMap[hash]; ok {
		if pc.routes == nil {
			pc.routes = make([]*Route, 0)
		}
		pc.routes = append(pc.routes, rte)
		return p, nil
	}
	var p Pluginer
	if pc.Mode == PluginModeInput {
		p, err = NewInputPlugin(ctx, rte)
	} else if pc.Mode == PluginModeOutput {
		p, err = NewOutputPlugin(ctx, rte)
	}
	if err != nil {
		return plugin, err
	}
	pm.pluginMap[hash] = p
	log.Debug().Msg(fmt.Sprintf("registering new %s %s plugin with hash %s", p.GetConfig().Type, p.GetConfig().Mode, hash))
	return p, nil
}

func (pm *DefaultIOPluginManager) WithdrawPlugin(ctx context.Context, rte *Route, plugin Pluginer) error {
	hash := plugin.Hash(ctx)
	if hash == "" {
		return new(EmptyPluginHashError)
	}
	pc := plugin.GetConfig()
	if p, ok := pm.pluginMap[hash]; ok {
		log.Debug().Msg(fmt.Sprintf("unregistering %s %s plugin with hash %s", p.GetConfig().Type, p.GetConfig().Mode, hash))
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
				log.Debug().Msg(fmt.Sprintf("stopped %s %s plugin with hash %s", p.GetConfig().Type, p.GetConfig().Mode, hash))
				delete(pm.pluginMap, hash)
			}
		}
		log.Debug().Msg(fmt.Sprintf("%s %s plugin route count %d %d", p.GetConfig().Type, p.GetConfig().Mode, p.GetRouteCount(), len(pc.routes)))
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
