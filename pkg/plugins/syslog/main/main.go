// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/xmidt-org/ears/pkg/plugins/syslog"
)

func main() {
	// required for `go build` to not fail
}

//go:generate ../../../../script/build-plugin.sh

var (
	Name       = "syslog"
	GitVersion = "v0.0.1"
	GitCommit  = ""
)

var Plugin, PluginErr = syslog.NewPluginVersion(Name, GitVersion, GitCommit)

// for golangci-lint
var _ = Plugin
var _ = PluginErr
