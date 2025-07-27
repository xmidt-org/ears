// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package dedup

import (
	lru "github.com/hashicorp/golang-lru"
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/tenant"
	"github.com/xorcare/pointer"
)

// Config can be passed into NewFilter() in order to configure
// the behavior of the sender.
type Config struct {
	CacheSize *int   `json:"cacheSize,omitempty"`
	Path      string `json:"path,omitempty"`
}

var DefaultConfig = Config{
	CacheSize: pointer.Int(1000),
	Path:      "",
}

type Filter struct {
	config   Config
	name     string
	plugin   string
	tid      tenant.Id
	lruCache *lru.Cache
	filter.MetricFilter
}
