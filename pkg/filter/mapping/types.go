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
