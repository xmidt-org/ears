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

package ttl

import (
	"github.com/xmidt-org/ears/pkg/tenant"
	"github.com/xorcare/pointer"
)

// Config can be passed into NewFilter() in order to configure
// the behavior of the sender.
type Config struct {
	Path       string `json:"path,omitempty"`
	Ttl        *int   `json:"ttl,omitempty"`        // ttl in MS
	NanoFactor *int   `json:"nanoFactor,omitempty"` // factor to convert timestamp to nano seconds
}

var DefaultConfig = Config{
	Path:       "",
	Ttl:        pointer.Int(1000 * 60 * 5), // 5 min in ms
	NanoFactor: pointer.Int(1000 * 1000),
}

type Filter struct {
	config Config
	name   string
	plugin string
	tid    tenant.Id
}
