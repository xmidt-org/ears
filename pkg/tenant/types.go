// Copyright 2021 Comcast Cable Communications Management, LLC
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

package tenant

import (
	"context"
	"encoding/base64"
)

const delimiter = "."

type Id struct {
	OrgId string `json:"orgId,omitempty"`
	AppId string `json:"appId,omitempty"`
}

func (id Id) Equal(tid Id) bool {
	return id.OrgId == tid.OrgId && id.AppId == tid.AppId
}

// Key A string representation of tenant id that can be used has a key
// for certain data structure like a map
func (id Id) Key() string {
	return base64.StdEncoding.EncodeToString([]byte(id.OrgId)) + delimiter +
		base64.StdEncoding.EncodeToString([]byte(id.AppId))
}

// KeyWithRoute A string representation of tenant id + route id that
// can be used has a key for certain data structure like a map
func (id Id) KeyWithRoute(routeId string) string {
	return id.Key() + delimiter + base64.StdEncoding.EncodeToString([]byte(routeId))
}

func (id Id) KeyWithFragment(fragmentId string) string {
	return id.OrgId + delimiter + id.AppId + delimiter + fragmentId
}

func (id Id) ToString() string {
	return "OrgId=" + id.OrgId + " AppId=" + id.AppId
}

type Config struct {
	Tenant    Id       `json:"tenant"`              // tenant id
	Quota     Quota    `json:"quota"`               // tenant quota
	ClientIds []string `json:"clientIds,omitempty"` // jwt subjects or client IDs
	Modified  int64    `json:"modified,omitempty"`  // last time when the tenant config is modified
}

type Quota struct {
	EventsPerSec int `json:"eventsPerSec"`
}

type TenantStorer interface {
	GetAllConfigs(ctx context.Context) ([]Config, error)
	GetConfig(ctx context.Context, id Id) (*Config, error)
	SetConfig(ctx context.Context, config Config) error
	DeleteConfig(ctx context.Context, id Id) error
}
