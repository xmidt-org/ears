// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/tenant"
	"github.com/xorcare/pointer"
)

// Config can be passed into NewFilter() in order to configure
// the behavior of the sender.
type Config struct {
	Path     string `json:"path,omitempty"`
	Tag      string `json:"tag,omitempty"`
	AsString *bool  `json:"asString,omitempty"`
	LogKey   string `json:"logKey,omitempty"`
}

var DefaultConfig = Config{
	Path:     "",
	Tag:      "",
	AsString: pointer.Bool(true),
	LogKey:   "payload",
}

type Filter struct {
	config Config
	name   string
	plugin string
	tid    tenant.Id
	filter.MetricFilter
}
