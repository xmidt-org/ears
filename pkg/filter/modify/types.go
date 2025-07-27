// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package modify

import (
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/tenant"
	"github.com/xorcare/pointer"
)

type Config struct {
	Path    string   `json:"path,omitempty"`
	Paths   []string `json:"paths,omitempty"`
	ToUpper *bool    `json:"toUpper,omitempty"`
	ToLower *bool    `json:"toLower,omitempty"`
}

var DefaultConfig = Config{
	Path:    "",
	Paths:   []string{},
	ToUpper: pointer.Bool(false),
	ToLower: pointer.Bool(false),
}

type Filter struct {
	config Config
	name   string
	plugin string
	tid    tenant.Id
	filter.MetricFilter
}
