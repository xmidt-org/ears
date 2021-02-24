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

package route

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sync"

	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/sender"
)

type Router interface {
	Run(ctx context.Context, r receiver.Receiver, f filter.Filterer, s sender.Sender) error
	Stop(ctx context.Context) error
}

type Route struct {
	sync.Mutex

	r receiver.Receiver
	f filter.Filterer
	s sender.Sender
}

type InvalidRouteError struct {
	Err error
}

type PluginConfig struct {
	Plugin string      `json:"plugin,omitempty"` // plugin or filter type, e.g. kafka, kds, sqs, webhook, filter
	Name   string      `json:"name,omitempty"`   // plugin label to allow multiple instances of otherwise identical plugin configurations
	Config interface{} `json:"config,omitempty"` // plugin specific configuration parameters
}

type Config struct {
	Id           string         `json:"id,omitempty"`           // route ID
	OrgId        string         `json:"orgId,omitempty"`        // org ID for quota and rate limiting
	AppId        string         `json:"appId,omitempty"`        // app ID for quota and rate limiting
	UserId       string         `json:"userId,omitempty"`       // user ID / author of route
	Name         string         `json:"name,omitempty"`         // optional unique name for route
	Receiver     PluginConfig   `json:"receiver,omitempty"`     // source plugin configuration
	Sender       PluginConfig   `json:"sender,omitempty"`       // destination plugin configuration
	FilterChain  []PluginConfig `json:"filterChain,omitempty"`  // filter chain configuration
	DeliveryMode string         `json:"deliveryMode,omitempty"` // possible values: fire_and_forget, at_least_once, exactly_once
	Debug        bool           `json:"debug,omitempty"`        // if true generate debug logs and metrics for events taking this route
	Created      int64          `json:"created,omitempty"`      // time on when route was created, in unix timestamp seconds
	Modified     int64          `json:"modified,omitempty"`     // last time when route was modified, in unix timestamp seconds
}

//Validate returns an error if the plugin config is invalid and nil otherwise
func (pc *PluginConfig) Validate(ctx context.Context) error {
	if pc.Plugin == "" {
		return errors.New("missing plugin type configuration")
	}
	if pc.Name != "" {
		validName := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*$`)
		if !validName.MatchString(pc.Name) {
			return errors.New("invalid plugin name " + pc.Name)
		}
	}
	return nil
}

//Validate returns an error if the route config is invalid and nil otherwise
func (rc *Config) Validate(ctx context.Context) error {
	var err error
	err = rc.Sender.Validate(ctx)
	if err != nil {
		return err
	}
	err = rc.Receiver.Validate(ctx)
	if err != nil {
		return err
	}
	if rc.FilterChain != nil {
		for _, f := range rc.FilterChain {
			err = f.Validate(ctx)
			if err != nil {
				return err
			}
		}
	}
	if rc.Id == "" {
		return errors.New("missing ID for plugin configuration")
	}
	if rc.Name != "" {
		validName := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*$`)
		if !validName.MatchString(rc.Name) {
			return errors.New("invalid route name " + rc.Name)
		}
	}
	if rc.OrgId == "" {
		return errors.New("missing org ID for plugin configuration")
	}
	if rc.AppId == "" {
		return errors.New("missing app ID for plugin configuration")
	}
	if rc.UserId == "" {
		return errors.New("missing user ID for plugin configuration")
	}
	return nil
}

//Hash returns the md5 hash of the plugin config
func (pc *PluginConfig) Hash(ctx context.Context) string {
	cfg := ""
	if pc.Config != nil {
		buf, _ := json.Marshal(pc.Config)
		if buf != nil {
			cfg = string(buf)
		}
	}
	str := pc.Name + pc.Plugin + cfg
	hash := fmt.Sprintf("%x", md5.Sum([]byte(str)))
	return hash
}

//Hash returns the md5 hash of a route config
func (pc *Config) Hash(ctx context.Context) string {
	// notably the route id is not part of the hash as the id might be the hash itself
	str := pc.OrgId + pc.AppId
	str += pc.Receiver.Hash(ctx)
	str += pc.Sender.Hash(ctx)
	if pc.FilterChain != nil {
		for _, f := range pc.FilterChain {
			str += f.Hash(ctx)
		}
	}
	hash := fmt.Sprintf("%x", md5.Sum([]byte(str)))
	return hash
}

//All route operations are synchronous. The storer should respect the cancellation
//from the context and cancel its operation gracefully when desired.
type RouteStorer interface {
	//If route is not found, the function should return a item not found error
	GetRoute(context.Context, string) (Config, error)
	GetAllRoutes(context.Context) ([]Config, error)

	//SetRoute will add the route if it new or update the route if
	//it is an existing one. It will also update the create time and
	//modified time of the route where appropriate.
	SetRoute(context.Context, Config) error

	SetRoutes(context.Context, []Config) error

	DeleteRoute(context.Context, string) error

	DeleteRoutes(context.Context, []string) error
}
