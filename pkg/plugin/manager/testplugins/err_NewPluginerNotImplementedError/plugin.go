// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package main

func main() {
	// required for `go build` to not fail
}

type plugin struct{}

var Plugin = plugin{}

// for golangci-lint
var _ = Plugin
