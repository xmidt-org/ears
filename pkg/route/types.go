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
	"sync"

	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/plugin"
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

type Config struct {
	Id           string          `json:"id,omitempty"`           // route ID
	OrgId        string          `json:"orgId,omitempty"`        // org ID for quota and rate limiting
	AppId        string          `json:"appId,omitempty"`        // app ID for quota and rate limiting
	UserId       string          `json:"userId,omitempty"`       // user ID / author of route
	Name         string          `json:"name,omitempty"`         // optional unique name for route
	Source       plugin.Config   `json:"source,omitempty"`       // source plugin configuration
	Destination  plugin.Config   `json:"destination,omitempty"`  // destination plugin configuration
	FilterChain  []plugin.Config `json:"filterChain,omitempty"`  // filter chain configuration
	DeliveryMode string          `json:"deliveryMode,omitempty"` // possible values: fire_and_forget, at_least_once, exactly_once
	Debug        bool            `json:"debug,omitempty"`        // if true generate debug logs and metrics for events taking this route
	Created      int64           `json:"ts,omitempty"`           // time on when route was created, in unix timestamp seconds
	Modified     int64           `json:"ts,omitempty"`           // last time when route was modified, in unix timestamp seconds
}

//All route operations are synchronous. The storer should respect the cancellation
//from the context and cancel its operation gracefully when desired.
type RouteStorer interface {
	GetRoute(context.Context, string) (*Config, error)
	GetAllRoutes(context.Context) ([]Config, error)

	//SetRoute will add the route if it new or update the route if
	//it is an existing one. It will also update the create time and
	//modified time of the route where appropriate.
	SetRoute(context.Context, Config) error

	SetRoutes(context.Context, []Config) error

	DeleteRoute(context.Context, string) error

	DeleteRoutes(context.Context, []string) error

	//For testing purpose only
	DeleteAllRoutes(ctx context.Context) error
}
