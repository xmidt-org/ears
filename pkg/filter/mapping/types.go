// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package mapping

import (
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/tenant"
)

// Config can be passed into NewFilter() in order to configure
// the behavior of the sender.
type Config struct {
	Map          []FromTo    `json:"map,omitempty"`
	Path         string      `json:"path,omitempty"`
	ArrayPath    string      `json:"arrayPath,omitempty"` // if arrayPath points to array, iterate over all elements and apply from and to paths relatively
	DefaultValue interface{} `json:"defaultValue,omitempty"`
}

type FromTo struct {
	From       interface{} `json:"from,omitempty"`
	To         interface{} `json:"to,omitempty"`
	Comparison *Comparison `json:"comparison,omitempty"`
}

type Comparison struct {
	Equal    []map[string]interface{} `json:"equal,omitempty"`
	NotEqual []map[string]interface{} `json:"notEqual,omitempty"`
	//GreaterThan []map[string]interface{} `json:"greaterThan,omitempty"`
	//LessThan    []map[string]interface{} `json:"lessThan,omitempty"`
}

var DefaultConfig = Config{
	Map: []FromTo{},
}

type Filter struct {
	config Config
	name   string
	plugin string
	tid    tenant.Id
	filter.MetricFilter
}
