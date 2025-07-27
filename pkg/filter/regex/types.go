// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package regex

import (
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/tenant"
)

// Config can be passed into NewFilter() in order to configure
// the behavior of the sender.
type Config struct {
	FromPath         string  `json:"fromPath,omitempty"`
	ToPath           string  `json:"toPath,omitempty"`
	Regex            string  `json:"regex,omitempty"`
	ReplaceAllString *string `json:"replaceAllString,omitempty"`
}

var DefaultConfig = Config{
	FromPath: "",
	ToPath:   "",
	Regex:    "^.*$",
}

type Filter struct {
	config Config
	name   string
	plugin string
	tid    tenant.Id
	filter.MetricFilter
}
