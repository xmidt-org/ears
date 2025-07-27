// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package ttl

import (
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/tenant"
	"github.com/xorcare/pointer"
	"go.opentelemetry.io/otel/metric"
)

// Config can be passed into NewFilter() in order to configure
// the behavior of the sender.
type Config struct {
	Path       string `json:"path,omitempty"`
	Ttl        *int   `json:"ttl,omitempty"`        // ttl in MS
	NanoFactor *int   `json:"nanoFactor,omitempty"` // factor to convert timestamp to nano seconds
	Disabled   *bool  `json:"disabled,omitempty"`
}

var DefaultConfig = Config{
	Path:       "",
	Ttl:        pointer.Int(1000 * 60 * 5), // 5 min in ms
	NanoFactor: pointer.Int(1000 * 1000),
	Disabled:   pointer.Bool(false),
}

type Filter struct {
	config                    Config
	name                      string
	plugin                    string
	tid                       tenant.Id
	eventTtlExpirationCounter metric.BoundInt64Counter
	filter.MetricFilter
}
