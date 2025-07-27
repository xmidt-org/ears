// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

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
