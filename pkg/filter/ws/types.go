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

package ws

import (
	"net/http"
	"sync"

	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
	"github.com/xorcare/pointer"
)

// Config can be passed into NewFilter() in order to configure
// the behavior of the sender.
type Config struct {
	ToPath                 string            `json:"toPath,omitempty"`
	FromPath               string            `json:"fromPath,omitempty"`
	Url                    string            `json:"url,omitempty"`
	UrlPath                string            `json:"urlPath,omitempty"`
	Method                 string            `json:"method,omitempty"`
	Body                   string            `json:"body,omitempty"`
	Headers                map[string]string `json:"headers,omitempty"`
	EmptyPathValueRequired *bool             `json:"emptyPathValueRequired,omitempty"`
	Auth                   *Auth             `json:"auth,omitempty"`
}

type Auth struct {
	Type         string   `json:"type,omitempty"`
	Username     string   `json:"username,omitempty"`     // basic auth
	Password     string   `json:"password,omitempty"`     // basic auth
	ClientID     string   `json:"clientId,omitempty"`     // oauth2
	ClientSecret string   `json:"clientSecret,omitempty"` // oauth2
	TokenURL     string   `json:"tokenUrl,omitempty"`     // oauth2
	Scopes       []string `json:"scopes,omitempty"`       // oauth2
	GrantType    string   `json:"grantType,omitempty"`    // oauth2
}

var DefaultConfig = Config{
	ToPath:                 "",
	FromPath:               "",
	Url:                    "",
	UrlPath:                "",
	Method:                 "GET",
	Body:                   "",
	Headers:                map[string]string{},
	EmptyPathValueRequired: pointer.Bool(false),
	Auth:                   &Auth{},
}

type Filter struct {
	sync.RWMutex
	config    Config
	name      string
	plugin    string
	tid       tenant.Id
	secrets   secret.Vault
	clients   map[string]*http.Client
	satTokens map[string]*SatToken
	filter.MetricFilter
}
