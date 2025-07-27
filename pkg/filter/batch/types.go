// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package batch

import (
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/tenant"
	"github.com/xorcare/pointer"
)

// Config can be passed into NewFilter() in order to configure
// the behavior of the sender.
type Config struct {
	BatchSize *int `json:"batchSize,omitempty"`
}

var DefaultConfig = Config{
	BatchSize: pointer.Int(1),
}

type Filter struct {
	config Config
	batch  []event.Event
	name   string
	plugin string
	tid    tenant.Id
	filter.MetricFilter
}
