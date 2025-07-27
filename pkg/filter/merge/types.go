// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package merge

import (
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/tenant"
)

// Config can be passed into NewFilter() in order to configure
// the behavior of the sender.
type Config struct {
	ToPath   string `json:"toPath,omitempty"`
	FromPath string `json:"fromPath,omitempty"`
}

var DefaultConfig = Config{
	ToPath:   "",
	FromPath: "",
}

type Filter struct {
	config Config
	name   string
	plugin string
	tid    tenant.Id
	filter.MetricFilter
}
