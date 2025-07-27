// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/xmidt-org/ears/pkg/plugins/ws"
)

func main() {
	// required for `go build` to not fail
}

//go:generate ../../../../script/build-plugin.sh

var (
	Name       = "ws"
	GitVersion = "v0.0.0"
	GitCommit  = ""
)

var Plugin, PluginErr = ws.NewPluginVersion(Name, GitVersion, GitCommit)

// for golangci-lint
var _ = Plugin
var _ = PluginErr
