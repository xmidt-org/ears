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

package pluginmanagerfx

import (
	"fmt"
	"github.com/rs/zerolog"
	p "github.com/xmidt-org/ears/internal/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/plugin/manager"
	"github.com/xmidt-org/ears/pkg/plugins/block"
	"github.com/xmidt-org/ears/pkg/plugins/debug"
	"github.com/xmidt-org/ears/pkg/plugins/dedup"
	"github.com/xmidt-org/ears/pkg/plugins/js"
	"github.com/xmidt-org/ears/pkg/plugins/kafka"
	"github.com/xmidt-org/ears/pkg/plugins/match"
	"github.com/xmidt-org/ears/pkg/plugins/pass"
	"github.com/xmidt-org/ears/pkg/plugins/redis"
	"github.com/xmidt-org/ears/pkg/plugins/split"
	"github.com/xmidt-org/ears/pkg/plugins/sqs"
	"github.com/xmidt-org/ears/pkg/plugins/transform"
	"github.com/xmidt-org/ears/pkg/plugins/unwrap"

	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		ProvidePluginManager,
	),
)

type PluginIn struct {
	fx.In

	Logger *zerolog.Logger
}

type PluginOut struct {
	fx.Out

	PluginManager p.Manager
}

func ProvidePluginManager(in PluginIn) (PluginOut, error) {
	out := PluginOut{}

	mgr, err := manager.New()
	if err != nil {
		return out, fmt.Errorf("could not provide plugin manager: %w", err)
	}

	// Go ahead and register some default plugins
	toArr := func(a ...interface{}) []interface{} { return a }

	defaultPlugins := []struct {
		name   string
		plugin pkgplugin.Pluginer
	}{
		{
			name:   "debug",
			plugin: toArr(debug.NewPluginVersion("debug", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "sqs",
			plugin: toArr(sqs.NewPluginVersion("sqs", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "redis",
			plugin: toArr(redis.NewPluginVersion("redis", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "kafka",
			plugin: toArr(kafka.NewPluginVersion("kafka", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "match",
			plugin: toArr(match.NewPluginVersion("match", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "pass",
			plugin: toArr(pass.NewPluginVersion("pass", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "js",
			plugin: toArr(js.NewPluginVersion("js", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "dedup",
			plugin: toArr(dedup.NewPluginVersion("dedup", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "block",
			plugin: toArr(block.NewPluginVersion("block", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "split",
			plugin: toArr(split.NewPluginVersion("split", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "unwrap",
			plugin: toArr(unwrap.NewPluginVersion("unwrap", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "transform",
			plugin: toArr(transform.NewPluginVersion("transform", "", ""))[0].(pkgplugin.Pluginer),
		},
	}

	for _, plug := range defaultPlugins {
		err = mgr.RegisterPlugin(plug.name, plug.plugin)
		if err != nil {
			return out, fmt.Errorf("could register %s plugin: %w", plug.name, err)
		}
	}

	m, err := p.NewManager(p.WithPluginManager(mgr), p.WithLogger(in.Logger))
	if err != nil {
		return out, fmt.Errorf("could not provide plugin manager: %w", err)
	}

	out.PluginManager = m

	return out, nil

}
