// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"github.com/xeipuuv/gojsonschema"
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/tenant"
)

// Config can be passed into NewFilter() in order to configure
// the behavior of the sender.
type Config struct {
	Path   string      `json:"path,omitempty"`
	Schema interface{} `json:"schema,omitempty"`
}

var DefaultConfig = Config{
	Path:   ".",
	Schema: map[string]interface{}{},
}

type Filter struct {
	config Config
	name   string
	plugin string
	tid    tenant.Id
	schema gojsonschema.JSONLoader
	filter.MetricFilter
}
