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

package modify

import (
	"github.com/xmidt-org/ears/pkg/tenant"
	"github.com/xorcare/pointer"
	"sync"
)

type Config struct {
	Path    string   `json:"path,omitempty"`
	Paths   []string `json:"paths,omitempty"`
	ToUpper *bool    `json:"toUpper,omitempty"`
	ToLower *bool    `json:"toLower,omitempty"`
}

var DefaultConfig = Config{
	Path:    "",
	Paths:   []string{},
	ToUpper: pointer.Bool(false),
	ToLower: pointer.Bool(false),
}

type Filter struct {
	sync.RWMutex
	config                        Config
	name                          string
	plugin                        string
	tid                           tenant.Id
	successCounter                int
	errorCounter                  int
	filterCounter                 int
	successVelocityCounter        int
	errorVelocityCounter          int
	filterVelocityCounter         int
	currentSuccessVelocityCounter int
	currentErrorVelocityCounter   int
	currentFilterVelocityCounter  int
	currentSec                    int64
}
