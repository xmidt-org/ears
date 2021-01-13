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

package main

import (
	"fmt"

	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
)

func main() {
	// required for `go build` to not fail
}

var Plugin = plugin{}

// for golangci-lint
var _ = Plugin

var _ pkgplugin.NewPluginerer = (*plugin)(nil)

// Plugin ============================================================

type plugin struct {
}

func (p *plugin) NewPluginer(config interface{}) (pkgplugin.Pluginer, error) {
	plug, _ := pkgplugin.NewPlugin(
		pkgplugin.WithName("name"),
		pkgplugin.WithVersion("version"),
		pkgplugin.WithCommitID("commitId"),
		pkgplugin.WithConfig(config),
	)
	return plug, &pkgplugin.InvalidConfigError{
		Err: fmt.Errorf("example config error"),
	}

}

func (p *plugin) PluginerHash(config interface{}) (string, error) {
	return "pluginerHash", nil
}
