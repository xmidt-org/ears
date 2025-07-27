// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package metric

import (
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel/metric"
)

// Config can be passed into NewFilter() in order to configure
// the behavior of the sender.
type Config struct {
	Name   string `json:"name,omitempty"`
	Prefix string `json:"prefix,omitempty"`
}

var DefaultConfig = Config{
	Name:   "generic",
	Prefix: "ears.custom.",
}

type Filter struct {
	config Config
	name   string
	plugin string
	tid    tenant.Id
	metric *metric.BoundInt64Counter
	filter.MetricFilter
}
