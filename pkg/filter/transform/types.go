// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package transform

import (
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/tenant"
)

// Config can be passed into NewFilter() in order to configure
// the behavior of the sender.
type Config struct {
	Transformation interface{} `json:"transformation,omitempty"`
	ToPath         string      `json:"toPath,omitempty"`
	FromPath       string      `json:"fromPath,omitempty"` // optional, if present apply transformation to sub event at path, if sub event is array apply transformation to all elements of array
}

var empty interface{}
var DefaultConfig = Config{
	Transformation: empty,
	ToPath:         "",
}

type Filter struct {
	config Config
	name   string
	plugin string
	tid    tenant.Id
	filter.MetricFilter
}
