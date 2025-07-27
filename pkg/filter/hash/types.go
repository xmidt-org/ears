// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package hash

import (
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/tenant"
)

// Config can be passed into NewFilter() in order to configure
// the behavior of the sender.
type Config struct {
	FromPath      string `json:"fromPath,omitempty"`
	From          string `json:"from,omitempty"`
	ToPath        string `json:"toPath,omitempty"`
	HashAlgorithm string `json:"hashAlgorithm,omitempty"`
	Key           string `json:"key,omitempty"`      // optional key for certain hash algorithms
	Encoding      string `json:"encoding,omitempty"` // optional encoding of hash, base64, hex etc.
}

var DefaultConfig = Config{
	FromPath:      "",
	From:          "",
	ToPath:        "",
	HashAlgorithm: "md5",
	Key:           "",
	Encoding:      "",
}

type Filter struct {
	config Config
	name   string
	plugin string
	tid    tenant.Id
	filter.MetricFilter
}
