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
	"github.com/xmidt-org/ears/internal/pkg/quota"
	"github.com/xmidt-org/ears/pkg/plugin/manager"
	"github.com/xmidt-org/ears/pkg/plugins/batch"
	"github.com/xmidt-org/ears/pkg/plugins/block"
	"github.com/xmidt-org/ears/pkg/plugins/debug"
	"github.com/xmidt-org/ears/pkg/plugins/decode"
	"github.com/xmidt-org/ears/pkg/plugins/dedup"
	"github.com/xmidt-org/ears/pkg/plugins/encode"
	"github.com/xmidt-org/ears/pkg/plugins/gears"
	"github.com/xmidt-org/ears/pkg/plugins/hash"
	"github.com/xmidt-org/ears/pkg/plugins/http"
	"github.com/xmidt-org/ears/pkg/plugins/js"
	"github.com/xmidt-org/ears/pkg/plugins/kafka"
	"github.com/xmidt-org/ears/pkg/plugins/kinesis"
	"github.com/xmidt-org/ears/pkg/plugins/log"
	"github.com/xmidt-org/ears/pkg/plugins/mapping"
	"github.com/xmidt-org/ears/pkg/plugins/match"
	"github.com/xmidt-org/ears/pkg/plugins/merge"
	"github.com/xmidt-org/ears/pkg/plugins/metric"
	"github.com/xmidt-org/ears/pkg/plugins/modify"
	"github.com/xmidt-org/ears/pkg/plugins/nop"
	"github.com/xmidt-org/ears/pkg/plugins/pass"
	"github.com/xmidt-org/ears/pkg/plugins/redis"
	"github.com/xmidt-org/ears/pkg/plugins/regex"
	"github.com/xmidt-org/ears/pkg/plugins/s3"
	"github.com/xmidt-org/ears/pkg/plugins/sample"
	"github.com/xmidt-org/ears/pkg/plugins/split"
	"github.com/xmidt-org/ears/pkg/plugins/sqs"
	"github.com/xmidt-org/ears/pkg/plugins/trace"
	"github.com/xmidt-org/ears/pkg/plugins/transform"
	"github.com/xmidt-org/ears/pkg/plugins/ttl"
	"github.com/xmidt-org/ears/pkg/plugins/unwrap"
	"github.com/xmidt-org/ears/pkg/plugins/validate"
	"github.com/xmidt-org/ears/pkg/plugins/ws"
	"github.com/xmidt-org/ears/pkg/secret"

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

	Logger       *zerolog.Logger
	QuotaManager *quota.QuotaManager
	Secrets      secret.Vault
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
			name:   "nop",
			plugin: toArr(nop.NewPluginVersion("nop", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "debug",
			plugin: toArr(debug.NewPluginVersion("debug", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "sqs",
			plugin: toArr(sqs.NewPluginVersion("sqs", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "s3",
			plugin: toArr(s3.NewPluginVersion("s3", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "kinesis",
			plugin: toArr(kinesis.NewPluginVersion("kinesis", "", ""))[0].(pkgplugin.Pluginer),
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
			name:   "gears",
			plugin: toArr(gears.NewPluginVersion("gears", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "match",
			plugin: toArr(match.NewPluginVersion("match", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "mapping",
			plugin: toArr(mapping.NewPluginVersion("mapping", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "pass",
			plugin: toArr(pass.NewPluginVersion("pass", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "log",
			plugin: toArr(log.NewPluginVersion("log", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "validate",
			plugin: toArr(validate.NewPluginVersion("validate", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "js",
			plugin: toArr(js.NewPluginVersion("js", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "ws",
			plugin: toArr(ws.NewPluginVersion("ws", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "sample",
			plugin: toArr(sample.NewPluginVersion("sample", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "metric",
			plugin: toArr(metric.NewPluginVersion("metric", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "dedup",
			plugin: toArr(dedup.NewPluginVersion("dedup", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "batch",
			plugin: toArr(batch.NewPluginVersion("batch", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "ttl",
			plugin: toArr(ttl.NewPluginVersion("ttl", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "modify",
			plugin: toArr(modify.NewPluginVersion("modify", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "trace",
			plugin: toArr(trace.NewPluginVersion("trace", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "decode",
			plugin: toArr(decode.NewPluginVersion("decode", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "encode",
			plugin: toArr(encode.NewPluginVersion("encode", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "regex",
			plugin: toArr(regex.NewPluginVersion("regex", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "hash",
			plugin: toArr(hash.NewPluginVersion("hash", "", ""))[0].(pkgplugin.Pluginer),
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
		{
			name:   "merge",
			plugin: toArr(merge.NewPluginVersion("merge", "", ""))[0].(pkgplugin.Pluginer),
		},
		{
			name:   "http",
			plugin: toArr(http.NewPluginVersion("http", "", ""))[0].(pkgplugin.Pluginer),
		},
	}

	for _, plug := range defaultPlugins {
		err = mgr.RegisterPlugin(plug.name, plug.plugin)
		if err != nil {
			return out, fmt.Errorf("could register %s plugin: %w", plug.name, err)
		}
	}

	m, err := p.NewManager(p.WithPluginManager(mgr),
		p.WithLogger(in.Logger),
		p.WithQuotaManager(in.QuotaManager),
		p.WithSecretVaults(in.Secrets))
	if err != nil {
		return out, fmt.Errorf("could not provide plugin manager: %w", err)
	}

	out.PluginManager = m

	return out, nil

}
